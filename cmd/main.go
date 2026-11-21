package main

import (
	"context"
	"log"

	"github.com/fury175/forge_api/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

const version = "0.0.1"

func main() {
	ctx := context.Background()

	cfg := config{
		addr: env.GetString("ADDR", ":3001"),
		env:  env.GetString("ENV", "development"),
	}

	conn, err := pgxpool.New(ctx, "user=admin host=localhost port=5432 dbname=forge_api sslmode=verify-full")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	app := application{config: cfg}

	log.Fatal(app.run(app.mount()))
}
