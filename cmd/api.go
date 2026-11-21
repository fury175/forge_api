package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	repo "github.com/fury175/forge_api/internal/adapters/postgresql/sqlc"
	"github.com/fury175/forge_api/internal/apperrors"
	"github.com/fury175/forge_api/internal/logger"
	"github.com/fury175/forge_api/internal/modules/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) mount() http.Handler {
	// chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// huma
	humaConfig := huma.DefaultConfig("Forge api", version)
	humaConfig.CreateHooks = nil
	humaConfig.DocsRenderer = huma.DocsRendererScalar
	humaConfig.Tags = []*huma.Tag{
		{Name: "ops"},
	}

	api := humachi.New(r, humaConfig)
	api.UseMiddleware(apperrors.Middleware)
	prefix := huma.NewGroup(api, "/v1")

	ops := huma.NewGroup(prefix, "")
	ops.UseMiddleware(app.basicAuthMiddleware(api))
	huma.Get(ops, "/health", app.healthCheckHandler, huma.OperationTags("ops"))

	authService := auth.NewService(repo.New(app.db), app.db, app.mailer, app.config.env, app.config.frontendUrl)
	authHandler := auth.NewHandler(authService)
	huma.Post(prefix, "/auth/sign-in", authHandler.LoginUser, huma.OperationTags("auth"))
	huma.Post(prefix, "/auth/sign-up", authHandler.Register, huma.OperationTags("auth"))

	return r
}

func (app *application) run(h http.Handler) error {
	svr := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Log.Infow("signal caught", "signal", s.String())

		shutdown <- svr.Shutdown(ctx)
	}()

	logger.Log.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	err := svr.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	logger.Log.Infow("server stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
