package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

const AccessTokenCookie = "__Host-access_token"
const RefreshTokenCookie = "__Host-refresh_token"
const SignedInCookie = "__Host-signed_in"

// SessionIsCurrent reports whether a token issued at issuedAt survives an account
// whose sessions were invalidated at validFrom.
//
// A JWT numeric date carries only whole seconds, so the cutoff is truncated to
// match. Keeping tokens issued during the cutoff's own second costs at most a
// second of revocation against a token that lives an hour, and is what lets a
// password change hand back a working session in the second it invalidates the
// old ones.
func SessionIsCurrent(issuedAt, validFrom time.Time) bool {
	return !issuedAt.Before(validFrom.Truncate(time.Second))
}

func AuthMiddleware(api huma.API, cfg *config.Config, queries *db.Queries) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		tokenCookie, err := huma.ReadCookie(ctx, AccessTokenCookie)
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

		wideLog.AddLogField(ctx.Context(), "userId", claims.Subject)

		// Access tokens are stateless, so this read is the only thing that can turn
		// away a signed-out, password-changed or deleted account before its token
		// expires on its own.
		sessionsValidFrom, err := queries.GetUserAuthState(ctx.Context(), claims.Subject)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
				return
			}

			wideLog.AddErrorField(ctx.Context(), err)
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Something went wrong on our end. Please try again.")
			return
		}

		if !SessionIsCurrent(claims.IssuedAt.Time, sessionsValidFrom) {
			wideLog.AddLogField(ctx.Context(), "staleSession", true)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx = huma.WithValue(ctx, ClaimsKey, claims)
		next(ctx)
	}
}
