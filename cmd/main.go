package main

import (
	"context"
	"log"
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
	defer pool.Close()

	app := &application{
		config: cfg,
		db:     pool,
	}

	h := app.mount()
	log.Fatal(app.run(h))
}
