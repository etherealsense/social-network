package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/etherealsense/social-network/pkg/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfg := config{
		env:  env.GetString("ENV"),
		addr: env.GetString("ADDR"),
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING"),
		},
		jwt: jwtConfig{
			secret:          env.GetString("JWT_SECRET"),
			accessTokenTTL:  time.Duration(env.GetInt("JWT_ACCESS_TOKEN_TTL")) * time.Hour,
			refreshTokenTTL: time.Duration(env.GetInt("JWT_REFRESH_TOKEN_TTL")) * time.Hour,
		},
		cors: corsConfig{
			origins: strings.Split(env.GetString("CORS_ORIGINS"), ","),
		},
	}

	var handler slog.Handler
	if cfg.env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	slog.SetDefault(slog.New(handler))

	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}

	app := &application{
		config: cfg,
		db:     pool,
	}

	h := app.mount()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.run(h); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	}()

	slog.Info("server started", "addr", cfg.addr, "env", cfg.env)

	<-quit
	slog.Info("shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
