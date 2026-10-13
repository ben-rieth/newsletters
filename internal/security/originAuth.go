package security

import (
	"crypto/subtle"
	"net/http"
)

const OriginAuthHeader = "X-Origin-Auth"

// RequireOriginSecret refuses requests that did not come through Cloudflare.
// Railway's edge is shared and routes by hostname, so anyone can reach the app
// there directly, skipping Cloudflare Access and choosing their own
// X-Forwarded-For. Only Cloudflare's transform rule attaches the secret.
//
// An empty secret disables the check, which config only allows outside prod.
func RequireOriginSecret(next http.Handler, secret string) http.Handler {
	if secret == "" {
		return next
	}

	want := []byte(secret)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get(OriginAuthHeader))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Nothing downstream needs it, and it must never end up in a log line.
		r.Header.Del(OriginAuthHeader)
		next.ServeHTTP(w, r)
	})
}
