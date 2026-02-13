package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/internal/chat"
	"github.com/etherealsense/social-network/internal/comment"
	"github.com/etherealsense/social-network/internal/follow"
	"github.com/etherealsense/social-network/internal/like"
	"github.com/etherealsense/social-network/internal/post"
	"github.com/etherealsense/social-network/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	config config
	db     *pgxpool.Pool
	server *http.Server
}

type config struct {
	env  string
	addr string
	db   dbConfig
	cors corsConfig
	auth auth.Config
}

type dbConfig struct {
	dsn string
}

type corsConfig struct {
	origins []string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.config.cors.origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(httprate.LimitByIP(100, time.Minute))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		repository := repo.New(app.db)

		authService := auth.NewService(repository)
		authHandler := auth.NewHandler(authService, app.config.auth)

		chatService := chat.NewService(repository)
		chatHub := chat.NewHub()
		chatHandler := chat.NewHandler(chatService, chatHub)

		// WebSocket route — no timeout or body limit middleware.
		r.Group(func(r chi.Router) {
			auth.RequireAuth(authHandler)(r)
			r.Get("/chats/{chat_id}/ws", chatHandler.HandleWebSocket)
		})

		// REST routes with timeout and body limit.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(time.Minute))
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
					next.ServeHTTP(w, r)
				})
			})

			r.Group(func(r chi.Router) {
				r.Use(httprate.LimitByIP(10, time.Minute))
				r.Post("/auth/register", authHandler.Register)
				r.Post("/auth/login", authHandler.Login)
			})

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Post("/auth/refresh", authHandler.Refresh)
				r.Post("/auth/logout", authHandler.Logout)
			})

			userService := user.NewService(repository)
			userHandler := user.NewHandler(userService)

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Get("/users/me", userHandler.GetMe)
				r.Put("/users/me", userHandler.UpdateUser)
			})

			postService := post.NewService(repository)
			postHandler := post.NewHandler(postService)
			r.Get("/posts/{id}", postHandler.GetPost)
			r.Get("/posts/user/{user_id}", postHandler.ListPostsByUserID)

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Post("/posts", postHandler.CreatePost)
				r.Put("/posts/{id}", postHandler.UpdatePost)
				r.Delete("/posts/{id}", postHandler.DeletePost)
			})

			commentService := comment.NewService(repository)
			commentHandler := comment.NewHandler(commentService)
			r.Get("/posts/{post_id}/comments", commentHandler.ListCommentsByPostID)
			r.Get("/comments/{id}", commentHandler.GetComment)

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Post("/posts/{post_id}/comments", commentHandler.CreateComment)
				r.Put("/comments/{id}", commentHandler.UpdateComment)
				r.Delete("/comments/{id}", commentHandler.DeleteComment)
			})

			followService := follow.NewService(repository)
			followHandler := follow.NewHandler(followService)
			r.Get("/users/{user_id}/followers", followHandler.ListFollowers)
			r.Get("/users/{user_id}/following", followHandler.ListFollowing)

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Post("/users/{user_id}/follow", followHandler.FollowUser)
				r.Delete("/users/{user_id}/follow", followHandler.UnfollowUser)
			})

			likeService := like.NewService(repository)
			likeHandler := like.NewHandler(likeService)
			r.Get("/posts/{post_id}/likes", likeHandler.ListLikesByPostID)

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Post("/posts/{post_id}/like", likeHandler.LikePost)
				r.Delete("/posts/{post_id}/like", likeHandler.UnlikePost)
			})

			r.Group(func(r chi.Router) {
				auth.RequireAuth(authHandler)(r)
				r.Post("/chats", chatHandler.CreateChat)
				r.Get("/chats", chatHandler.ListChats)
				r.Get("/chats/{chat_id}/participants", chatHandler.ListParticipants)
				r.Get("/chats/{chat_id}/messages", chatHandler.ListMessages)
			})
		})
	})

	return r
}

func (app *application) run(h http.Handler) error {
	app.server = &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	slog.Info("server has started", "addr", app.config.addr)

	return app.server.ListenAndServe()
}

func (app *application) shutdown(ctx context.Context) error {
	slog.Info("shutting down http server...")
	if err := app.server.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("closing database connection...")
	app.db.Close()

	return nil
}
