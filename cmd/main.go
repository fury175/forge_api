package main

import (
	"log"

	"github.com/fury175/forge_api/internal/env"
	_ "github.com/joho/godotenv/autoload"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":3001"),
		env:  env.GetString("ENV", "development"),
	}

	app := application{config: cfg}

	log.Fatal(app.run(app.mount()))
}
