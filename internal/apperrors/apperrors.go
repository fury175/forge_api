package apperrors

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/fury175/forge_api/internal/logger"
	"go.uber.org/zap"
)

type humaCtxKey struct{}

func GetHumaCtx(ctx context.Context) huma.Context {
	v, _ := ctx.Value(humaCtxKey{}).(huma.Context)
	return v
}

func Middleware(humaCtx huma.Context, next func(huma.Context)) {
	next(huma.WithValue(humaCtx, humaCtxKey{}, humaCtx))
}

func log() *zap.SugaredLogger {
	return logger.Log.WithOptions(zap.AddCallerSkip(1))
}

func logFields(ctx context.Context, error error) []any {
	humaCtx := GetHumaCtx(ctx)
	return []any{
		"method", humaCtx.Method(),
		"url", humaCtx.URL().Path,
		"error", error.Error(),
	}
}

func InternalServerError(ctx context.Context, err error) error {
	log().Errorw("Internal server error", logFields(ctx, err)...)
	return huma.Error500InternalServerError("something went wrong, try again")
}

func NotFoundError(ctx context.Context, err error) error {
	log().Warnw("not found", logFields(ctx, err)...)
	return huma.Error404NotFound(err.Error())
}

func ConflictError(ctx context.Context, err error) error {
	log().Warnw("conflict", logFields(ctx, err)...)
	return huma.Error409Conflict(err.Error())
}

func BadRequestError(ctx context.Context, err error) error {
	log().Warnw("bad request", logFields(ctx, err)...)
	return huma.Error400BadRequest(err.Error())
}

func UnauthorizedAccessError(ctx context.Context, err error) error {
	log().Warnw("unauthorized access", logFields(ctx, err)...)
	return huma.Error401Unauthorized(err.Error())
}
