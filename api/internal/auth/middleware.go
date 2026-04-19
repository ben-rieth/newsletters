package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
)

var invalidBearerError = errors.New("Invalid Authorization Header")

func GetTokenFromAuthorizationHeader(header string) (string, error) {
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return "", invalidBearerError
	}

	token := strings.TrimPrefix(header, "Bearer ")
	return token, nil
}

func AuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		cfg := config.Load()

		authHeader := ctx.Header("Authorization")
		tokenStr, err := GetTokenFromAuthorizationHeader(authHeader)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Proper authorization header not included")
			return
		}

		claims, err := ParseToken(tokenStr, cfg.JWTSecret)
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
