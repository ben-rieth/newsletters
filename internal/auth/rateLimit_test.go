package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

// Aliased so the embedded field is not named Context, which would shadow the
// interface's own Context() method.
type embeddedHumaContext = huma.Context

type fakeContext struct {
	embeddedHumaContext
	remoteAddr string
	headers    http.Header
}

func (c fakeContext) RemoteAddr() string { return c.remoteAddr }

func (c fakeContext) Header(name string) string { return c.headers.Get(name) }

func ctxWith(remoteAddr, cfConnectingIP string) fakeContext {
	headers := http.Header{}
	if cfConnectingIP != "" {
		headers.Set(cloudflareClientIPHeader, cfConnectingIP)
	}

	return fakeContext{remoteAddr: remoteAddr, headers: headers}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name            string
		remoteAddr      string
		cfConnectingIP  string
		trustCloudflare bool
		want            string
	}{
		{
			name:            "untrusted ignores the header, since anyone can send it",
			remoteAddr:      "10.0.0.1:5000",
			cfConnectingIP:  "203.0.113.9",
			trustCloudflare: false,
			want:            "10.0.0.1",
		},
		{
			name:            "trusted takes the client from the header",
			remoteAddr:      "100.64.0.13:13040",
			cfConnectingIP:  "203.0.113.9",
			trustCloudflare: true,
			want:            "203.0.113.9",
		},
		{
			name:            "trusted accepts an ipv6 client",
			remoteAddr:      "100.64.0.13:13040",
			cfConnectingIP:  "2001:db8::1",
			trustCloudflare: true,
			want:            "2001:db8::1",
		},
		{
			name:            "a missing header falls back to the peer",
			remoteAddr:      "10.0.0.1:5000",
			trustCloudflare: true,
			want:            "10.0.0.1",
		},
		{
			name:            "a garbage header falls back to the peer rather than trusting it",
			remoteAddr:      "10.0.0.1:5000",
			cfConnectingIP:  "not-an-ip",
			trustCloudflare: true,
			want:            "10.0.0.1",
		},
		{
			name:            "an ipv4 mapped client is unmapped so it buckets with its ipv4 form",
			remoteAddr:      "10.0.0.1:5000",
			cfConnectingIP:  "::ffff:203.0.113.9",
			trustCloudflare: true,
			want:            "203.0.113.9",
		},
		{
			name:            "an ipv6 peer is used when untrusted",
			remoteAddr:      "[2001:db8::1]:5000",
			trustCloudflare: false,
			want:            "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clientIP(ctxWith(tt.remoteAddr, tt.cfConnectingIP), tt.trustCloudflare)
			if err != nil {
				t.Fatalf("clientIP returned error: %v", err)
			}

			if got.String() != tt.want {
				t.Errorf("clientIP = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// Railway's edge rewrites X-Forwarded-For to name Cloudflare's egress and its
// own hop, so trusting any entry of it would bucket visitors by Cloudflare server.
func TestClientIPIgnoresForwardedFor(t *testing.T) {
	ctx := ctxWith("100.64.0.13:13040", "203.0.113.9")
	ctx.headers.Set("X-Forwarded-For", "198.51.100.48, 192.0.2.194")

	got, err := clientIP(ctx, true)
	if err != nil {
		t.Fatalf("clientIP returned error: %v", err)
	}

	if got.String() != "203.0.113.9" {
		t.Errorf("clientIP = %q, want 203.0.113.9", got.String())
	}
}

func TestRateLimitKey(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{name: "ipv4 buckets by address", addr: "203.0.113.9", want: "203.0.113.9"},
		{name: "ipv6 buckets by /64", addr: "2001:db8:1:2::1", want: "2001:db8:1:2::/64"},
		{name: "ipv4 mapped ipv6 buckets as ipv4", addr: "::ffff:203.0.113.9", want: "203.0.113.9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rateLimitKey(netip.MustParseAddr(tt.addr)); got != tt.want {
				t.Errorf("rateLimitKey(%s) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

func TestRateLimitKeyCollapsesIPv6Allocation(t *testing.T) {
	first := rateLimitKey(netip.MustParseAddr("2001:db8:1:2::1"))
	second := rateLimitKey(netip.MustParseAddr("2001:db8:1:2:dead:beef:cafe:9999"))

	if first != second {
		t.Errorf("addresses in one /64 got different keys: %q and %q", first, second)
	}

	other := rateLimitKey(netip.MustParseAddr("2001:db8:1:3::1"))
	if first == other {
		t.Errorf("addresses in different /64s shared key %q", first)
	}
}

func TestGetLimiterEvictsOldestPastCeiling(t *testing.T) {
	limiter := KeyedLimiter{
		limiters: make(map[string]*Limiter),
		limit:    1,
		burst:    1,
	}

	for i := range maxLimiters {
		limiter.GetLimiter(netip.AddrFrom4([4]byte{10, byte(i >> 16), byte(i >> 8), byte(i)}).String())
	}

	if len(limiter.limiters) != maxLimiters {
		t.Fatalf("expected map to fill to %d, got %d", maxLimiters, len(limiter.limiters))
	}

	limiter.GetLimiter("203.0.113.9")

	if len(limiter.limiters) > maxLimiters {
		t.Errorf("map grew past ceiling to %d", len(limiter.limiters))
	}

	if _, ok := limiter.limiters["203.0.113.9"]; !ok {
		t.Error("new client was not admitted")
	}
}

func newTestLimiter(burst int) *KeyedLimiter {
	return &KeyedLimiter{
		limiters: make(map[string]*Limiter),
		limit:    0,
		burst:    burst,
	}
}

// The sign-in budget is keyed by account and address together, so a stranger
// spending theirs must leave the owner's untouched.
func TestAllowKeepsKeysIndependent(t *testing.T) {
	limiter := newTestLimiter(2)

	for range 2 {
		limiter.Allow("account|203.0.113.9")
	}

	if limiter.Allow("account|203.0.113.9") {
		t.Error("exhausted key was allowed")
	}

	if !limiter.Allow("account|198.51.100.7") {
		t.Error("same account from another address was refused")
	}
}

func TestResetRestoresFullBurst(t *testing.T) {
	limiter := newTestLimiter(2)

	for range 2 {
		limiter.Allow("key")
	}
	limiter.Reset("key")

	for i := range 2 {
		if !limiter.Allow("key") {
			t.Fatalf("attempt %d after reset was refused", i+1)
		}
	}
}

type middlewareProbe struct {
	handler   http.Handler
	calls     int
	clientKey string
}

// The limiter never refills, so a test controls exactly how many requests pass.
func newMiddlewareProbe(t *testing.T, burst int) *middlewareProbe {
	handler, api := humatest.New(t)
	api.UseMiddleware(NewRateLimitMiddleware(t.Context(), api, &config.Config{}, 0, burst))

	probe := &middlewareProbe{handler: handler}
	huma.Get(api, "/probe", func(ctx context.Context, _ *struct{}) (*struct{}, error) {
		probe.calls++
		probe.clientKey, _ = ClientKeyFromContext(ctx)
		return nil, nil
	})

	return probe
}

func (p *middlewareProbe) serve(remoteAddr string) int {
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.RemoteAddr = remoteAddr

	rec := httptest.NewRecorder()
	p.handler.ServeHTTP(rec, req)
	return rec.Code
}

func TestRateLimitMiddlewareExposesClientKey(t *testing.T) {
	probe := newMiddlewareProbe(t, 1)

	probe.serve("[2001:db8:1:2::1]:5000")

	if probe.clientKey != "2001:db8:1:2::/64" {
		t.Errorf("client key = %q, want the bucketed 2001:db8:1:2::/64", probe.clientKey)
	}
}

func TestRateLimitMiddlewareRefusesPastBurst(t *testing.T) {
	probe := newMiddlewareProbe(t, 2)

	for i := range 2 {
		if code := probe.serve("203.0.113.9:5000"); code != http.StatusNoContent {
			t.Fatalf("request %d within burst = %d, want %d", i+1, code, http.StatusNoContent)
		}
	}

	if code := probe.serve("203.0.113.9:5000"); code != http.StatusTooManyRequests {
		t.Errorf("request past burst = %d, want %d", code, http.StatusTooManyRequests)
	}

	if probe.calls != 2 {
		t.Errorf("handler ran %d times, want 2", probe.calls)
	}

	if code := probe.serve("198.51.100.7:5000"); code != http.StatusNoContent {
		t.Errorf("another address was refused with %d", code)
	}
}

func TestRateLimitMiddlewareFailsClosedOnUnparsablePeer(t *testing.T) {
	probe := newMiddlewareProbe(t, 1)

	if code := probe.serve("not-an-address"); code != http.StatusInternalServerError {
		t.Errorf("unparsable peer = %d, want %d", code, http.StatusInternalServerError)
	}

	if probe.calls != 0 {
		t.Error("handler ran for a request with no usable client address")
	}
}

// A slow limiter like the daily email budget can sit idle for hours while still
// spent, so cleanup must judge buckets by their tokens rather than their age.
func TestCleanUpKeepsSpentBucketsAndDropsFullOnes(t *testing.T) {
	limiter := newTestLimiter(2)

	limiter.GetLimiter("idle")
	limiter.Allow("spent")
	limiter.limiters["spent"].lastSeen = time.Now().Add(-48 * time.Hour)

	limiter.cleanUp()

	if _, ok := limiter.limiters["idle"]; ok {
		t.Error("full bucket survived cleanup")
	}

	if _, ok := limiter.limiters["spent"]; !ok {
		t.Error("spent bucket was dropped, which would refund its budget")
	}
}
