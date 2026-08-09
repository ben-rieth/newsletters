package auth

import (
	"errors"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
)

func AuthMiddleware(api huma.API, cfg *config.Config) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		tokenCookie, err := huma.ReadCookie(ctx, "access_token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				huma.WriteErr(api, ctx, http.StatusUnauthorized, "Access token not included with request")
				return
			}

			wideLog.AddErrorField(ctx.Context(), err)
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Something went wrong on our end. Please try again.")
			return
		}

		claims, err := ParseToken(tokenCookie.Value, cfg.JWTSecret)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if claims != nil {
			wideLog.AddLogField(ctx.Context(), "userId", claims.Subject)
		}

		ctx = huma.WithValue(ctx, ClaimsKey, claims)
		next(ctx)
	}
}
