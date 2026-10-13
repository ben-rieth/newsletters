package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testOriginSecret = "0123456789abcdef0123456789abcdef"

type originProbe struct {
	calls      int
	seenHeader string
}

func (p *originProbe) serve(t *testing.T, secret string, header *string) int {
	t.Helper()

	handler := RequireOriginSecret(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.calls++
		p.seenHeader = r.Header.Get(OriginAuthHeader)
		w.WriteHeader(http.StatusOK)
	}), secret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != nil {
		req.Header.Set(OriginAuthHeader, *header)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec.Code
}

func ptr(s string) *string { return &s }

func TestRequireOriginSecretRefusesRequestsWithoutIt(t *testing.T) {
	tests := []struct {
		name   string
		header *string
	}{
		{name: "missing header", header: nil},
		{name: "empty header", header: ptr("")},
		{name: "wrong secret", header: ptr("fedcba9876543210fedcba9876543210")},
		{name: "prefix of the secret", header: ptr(testOriginSecret[:16])},
		{name: "secret with extra bytes", header: ptr(testOriginSecret + "x")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := &originProbe{}

			if code := probe.serve(t, testOriginSecret, tt.header); code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", code, http.StatusForbidden)
			}

			if probe.calls != 0 {
				t.Error("request reached the app without the origin secret")
			}
		})
	}
}

func TestRequireOriginSecretPassesCloudflareRequests(t *testing.T) {
	probe := &originProbe{}

	if code := probe.serve(t, testOriginSecret, ptr(testOriginSecret)); code != http.StatusOK {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}

	if probe.calls != 1 {
		t.Errorf("handler ran %d times, want 1", probe.calls)
	}

	if probe.seenHeader != "" {
		t.Error("secret was passed on to the app, where it could be logged")
	}
}

// Dev has no Cloudflare in front of it, so an unset secret must not lock it out.
func TestRequireOriginSecretIsDisabledWithoutASecret(t *testing.T) {
	probe := &originProbe{}

	if code := probe.serve(t, "", nil); code != http.StatusOK {
		t.Errorf("status = %d, want %d", code, http.StatusOK)
	}
}
