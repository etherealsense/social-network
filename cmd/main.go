package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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
		log.Panic(err)
	}

	cfg := config{
		env:  env.GetString("ENV"),
		addr: env.GetString("ADDR"),
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING"),
		},
		jwtSecret:          env.GetString("JWT_SECRET"),
		jwtAccessTokenTTL:  time.Duration(env.GetInt("JWT_ACCESS_TOKEN_TTL")) * time.Hour,
		jwtRefreshTokenTTL: time.Duration(env.GetInt("JWT_REFRESH_TOKEN_TTL")) * time.Hour,
	}

	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		log.Panic(err)
	}

	app := &application{
		config: cfg,
		db:     pool,
	}

	h := app.mount()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.run(h); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	log.Println("server started")

	<-quit
	log.Println("shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("server stopped")
}
