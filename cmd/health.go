package main

import "context"

type HealthBody struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Env     string `json:"env"`
}

type Health struct {
	Body HealthBody
}

func (app *application) healthCheckHandler(ctx context.Context, input *struct{}) (*Health, error) {
	return &Health{Body: HealthBody{Status: "ok", Success: true, Env: app.config.env}}, nil
}
