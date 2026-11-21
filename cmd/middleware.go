package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func (app *application) basicAuthMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")

		if authHeader == "" {
			app.writeUnauthorized(api, ctx, "authorization header is missing")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			app.writeUnauthorized(api, ctx, "authorization header is malformed")
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			app.writeUnauthorized(api, ctx, "invalid credentials")
			return
		}
		creds := strings.SplitN(string(decoded), ":", 2)
		if len(creds) != 2 || creds[0] != app.config.auth.basic.username || creds[1] != app.config.auth.basic.password {
			app.writeUnauthorized(api, ctx, "invalid credentials")
			return
		}
		next(ctx)
	}
}

func (app *application) writeUnauthorized(api huma.API, ctx huma.Context, msg string) {
	ctx.SetHeader("WWW-Authenticate", `Basic realm="restricted"`)
	huma.WriteErr(api, ctx, http.StatusUnauthorized, msg)
}
