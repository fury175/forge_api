package main

import (
	"time"

	"github.com/fury175/forge_api/internal/mailer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	config config
	db     *pgxpool.Pool
	mailer mailer.Client
}

type config struct {
	addr        string
	env         string
	mail        mailConfig
	auth        authConfig
	frontendUrl string
}

type authConfig struct {
	basic basicAuthConfig
}

type basicAuthConfig struct {
	username string
	password string
}

type mailConfig struct {
	fromEmail string
	mailTrap  mailTrapConfig
	exp       time.Duration
}

type mailTrapConfig struct {
	apiKey          string
	sandboxUsername string
}
