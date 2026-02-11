package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/internal/comment"
	"github.com/etherealsense/social-network/internal/follow"
	"github.com/etherealsense/social-network/internal/like"
	"github.com/etherealsense/social-network/internal/post"
	"github.com/etherealsense/social-network/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	config config
	db     *pgxpool.Pool
	server *http.Server
}

type config struct {
	env                string
	addr               string
	db                 dbConfig
	jwtSecret          string
	jwtAccessTokenTTL  time.Duration
	jwtRefreshTokenTTL time.Duration
}

type dbConfig struct {
	dsn string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(time.Minute))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		repository := repo.New(app.db)

		jwtAuth := auth.NewJWTAuth(app.config.jwtSecret, app.config.jwtAccessTokenTTL, app.config.jwtRefreshTokenTTL)
		authService := auth.NewService(repository)
		cookieConfig := auth.CookieConfig{
			Secure:   app.config.env == "production",
			SameSite: http.SameSiteLaxMode,
		}
		authHandler := auth.NewHandler(authService, jwtAuth, cookieConfig)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth.Verifier(jwtAuth))
			r.Use(auth.Authenticator(jwtAuth))
			r.Use(auth.ExtractUserID)
			r.Post("/auth/refresh", authHandler.Refresh)
			r.Post("/auth/logout", authHandler.Logout)
		})

		userService := user.NewService(repository)
		userHandler := user.NewHandler(userService)

		r.Group(func(r chi.Router) {
			r.Use(auth.Verifier(jwtAuth))
			r.Use(auth.Authenticator(jwtAuth))
			r.Use(auth.ExtractUserID)
			r.Get("/users/me", userHandler.GetMe)
			r.Put("/users/me", userHandler.UpdateUser)
		})

		postService := post.NewService(repository)
		postHandler := post.NewHandler(postService)
		r.Get("/posts/{id}", postHandler.GetPost)
		r.Get("/posts/user/{user_id}", postHandler.ListPostsByUserID)

		r.Group(func(r chi.Router) {
			r.Use(auth.Verifier(jwtAuth))
			r.Use(auth.Authenticator(jwtAuth))
			r.Use(auth.ExtractUserID)
			r.Post("/posts", postHandler.CreatePost)
			r.Put("/posts/{id}", postHandler.UpdatePost)
			r.Delete("/posts/{id}", postHandler.DeletePost)
		})

		commentService := comment.NewService(repository)
		commentHandler := comment.NewHandler(commentService)
		r.Get("/posts/{post_id}/comments", commentHandler.ListCommentsByPostID)
		r.Get("/comments/{id}", commentHandler.GetComment)

		r.Group(func(r chi.Router) {
			r.Use(auth.Verifier(jwtAuth))
			r.Use(auth.Authenticator(jwtAuth))
			r.Use(auth.ExtractUserID)
			r.Post("/posts/{post_id}/comments", commentHandler.CreateComment)
			r.Put("/comments/{id}", commentHandler.UpdateComment)
			r.Delete("/comments/{id}", commentHandler.DeleteComment)
		})

		followService := follow.NewService(repository)
		followHandler := follow.NewHandler(followService)
		r.Get("/users/{user_id}/followers", followHandler.ListFollowers)
		r.Get("/users/{user_id}/following", followHandler.ListFollowing)

		r.Group(func(r chi.Router) {
			r.Use(auth.Verifier(jwtAuth))
			r.Use(auth.Authenticator(jwtAuth))
			r.Use(auth.ExtractUserID)
			r.Post("/users/{user_id}/follow", followHandler.FollowUser)
			r.Delete("/users/{user_id}/follow", followHandler.UnfollowUser)
		})

		likeService := like.NewService(repository)
		likeHandler := like.NewHandler(likeService)
		r.Get("/posts/{post_id}/likes", likeHandler.ListLikesByPostID)

		r.Group(func(r chi.Router) {
			r.Use(auth.Verifier(jwtAuth))
			r.Use(auth.Authenticator(jwtAuth))
			r.Use(auth.ExtractUserID)
			r.Post("/posts/{post_id}/like", likeHandler.LikePost)
			r.Delete("/posts/{post_id}/like", likeHandler.UnlikePost)
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
