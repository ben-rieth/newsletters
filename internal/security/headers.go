package security

import "net/http"

// The SPA and the API are one origin and every asset ships in the binary, so
// nothing legitimate is ever fetched from another host.
//
// style-src allows inline because Base UI positions floating elements and sonner
// injects its stylesheet through the style attribute at runtime.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data:; " +
	"font-src 'self'; " +
	"connect-src 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"object-src 'none'"

// No preload: the list is slow to leave and would commit every current and future
// subdomain of the apex to HTTPS, including ones this app knows nothing about.
const strictTransportSecurity = "max-age=31536000; includeSubDomains"

// Headers applies the response headers the browser needs to enforce origin
// boundaries this server cannot enforce on its own.
//
// CSP and HSTS are production-only. Both assume the served frontend is the built
// bundle over HTTPS; under `make dev` the Go server reverse-proxies the Vite dev
// server, whose HMR client would trip the policy on rules production never
// evaluates. Run the binary with ENVIRONMENT=prod to exercise them locally.
func Headers(next http.Handler, prod bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		if prod {
			header.Set("Content-Security-Policy", contentSecurityPolicy)
			header.Set("Strict-Transport-Security", strictTransportSecurity)
		}

		next.ServeHTTP(w, r)
	})
}
