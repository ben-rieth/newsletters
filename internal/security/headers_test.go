package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveHeaders(t *testing.T, prod bool) http.Header {
	t.Helper()

	rec := httptest.NewRecorder()
	handler := Headers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), prod)

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	return rec.Result().Header
}

func TestHeadersAlwaysSetOriginProtections(t *testing.T) {
	for _, prod := range []bool{true, false} {
		header := serveHeaders(t, prod)

		if got := header.Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("prod=%v: X-Content-Type-Options = %q, want nosniff", prod, got)
		}

		if header.Get("Referrer-Policy") == "" {
			t.Errorf("prod=%v: Referrer-Policy not set", prod)
		}
	}
}

func TestHeadersEnforcePolicyOnlyInProd(t *testing.T) {
	prodHeader := serveHeaders(t, true)

	if got := prodHeader.Get("Content-Security-Policy"); got != contentSecurityPolicy {
		t.Errorf("Content-Security-Policy = %q, want the configured policy", got)
	}

	if got := prodHeader.Get("Strict-Transport-Security"); got != strictTransportSecurity {
		t.Errorf("Strict-Transport-Security = %q, want %q", got, strictTransportSecurity)
	}

	devHeader := serveHeaders(t, false)

	if got := devHeader.Get("Content-Security-Policy"); got != "" {
		t.Errorf("dev Content-Security-Policy = %q, want empty", got)
	}

	if got := devHeader.Get("Strict-Transport-Security"); got != "" {
		t.Errorf("dev Strict-Transport-Security = %q, want empty", got)
	}
}
