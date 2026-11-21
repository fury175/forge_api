package main

import (
	"context"
	"log"
	"time"

	"github.com/fury175/forge_api/internal/env"
	"github.com/fury175/forge_api/internal/mailer"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

const version = "0.0.1"

func main() {
	ctx := context.Background()

	cfg := config{
		addr:        env.GetString("ADDR", ":3001"),
		env:         env.GetString("ENV", "development"),
		frontendUrl: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		auth: authConfig{
			basic: basicAuthConfig{
				username: env.GetString("USERNAME", "admin"),
				password: env.GetString("PASSWORD", "pass"),
			},
		},
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: "no-reply@forgeapi.dev",
			mailTrap: mailTrapConfig{
				apiKey:          env.GetString("MAILTRAP_API_KEY", ""),
				sandboxUsername: env.GetString("SANDBOX_USERNAME", ""),
			},
		},
	}

	dbUrl := env.GetString("GOOSE_DBSTRING", "host=localhost user=admin password=pass port=5437 dbname=chill sslmode=disable")
	conn, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	mailer, err := mailer.NewMailTrapClient(cfg.mail.fromEmail, cfg.mail.mailTrap.apiKey, cfg.mail.mailTrap.sandboxUsername)
	if err != nil {
		panic(err)
	}

	app := application{
		config: cfg,
		db:     conn,
		mailer: mailer,
	}

	log.Fatal(app.run(app.mount()))
}
