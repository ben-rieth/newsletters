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

type header struct{ name, value string }

// Headers are a slice because a request can repeat X-Forwarded-For, and the
// order of the repeats decides which hop is trusted.
type fakeContext struct {
	embeddedHumaContext
	remoteAddr string
	headers    []header
}

func (c fakeContext) RemoteAddr() string { return c.remoteAddr }

func (c fakeContext) EachHeader(cb func(name, value string)) {
	for _, h := range c.headers {
		cb(h.name, h.value)
	}
}

func ctxWith(remoteAddr string, forwardedFor ...string) fakeContext {
	var headers []header
	for _, value := range forwardedFor {
		if value != "" {
			headers = append(headers, header{"X-Forwarded-For", value})
		}
	}

	return fakeContext{remoteAddr: remoteAddr, headers: headers}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name              string
		remoteAddr        string
		forwardedFor      string
		trustedProxyCount int
		want              string
	}{
		{
			name:              "no proxies trusted ignores the header entirely",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "1.2.3.4",
			trustedProxyCount: 0,
			want:              "10.0.0.1",
		},
		{
			name:              "one proxy takes the last hop",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "203.0.113.9",
			trustedProxyCount: 1,
			want:              "203.0.113.9",
		},
		{
			name:              "two proxies take the second to last hop",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "203.0.113.9, 172.16.0.1",
			trustedProxyCount: 2,
			want:              "203.0.113.9",
		},
		{
			name:              "a spoofed hop is left of the trusted window and ignored",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "1.2.3.4, 203.0.113.9, 172.16.0.1",
			trustedProxyCount: 2,
			want:              "203.0.113.9",
		},
		{
			name:              "a chain shorter than the trusted count falls back to the peer",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "203.0.113.9",
			trustedProxyCount: 2,
			want:              "10.0.0.1",
		},
		{
			name:              "a missing chain falls back to the peer",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "",
			trustedProxyCount: 2,
			want:              "10.0.0.1",
		},
		{
			name:              "a garbage hop falls back to the peer rather than trusting it",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "not-an-ip, 172.16.0.1",
			trustedProxyCount: 2,
			want:              "10.0.0.1",
		},
		{
			name:              "an ipv6 peer is used when no proxies are trusted",
			remoteAddr:        "[2001:db8::1]:5000",
			trustedProxyCount: 0,
			want:              "2001:db8::1",
		},
		{
			name:              "an ipv4 mapped hop is unmapped so it buckets with its ipv4 form",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "::ffff:203.0.113.9, 172.16.0.1",
			trustedProxyCount: 2,
			want:              "203.0.113.9",
		},
		{
			name:              "empty hops are skipped rather than counted",
			remoteAddr:        "10.0.0.1:5000",
			forwardedFor:      "203.0.113.9,, 172.16.0.1,",
			trustedProxyCount: 2,
			want:              "203.0.113.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clientIP(ctxWith(tt.remoteAddr, tt.forwardedFor), tt.trustedProxyCount)
			if err != nil {
				t.Fatalf("clientIP returned error: %v", err)
			}

			if got.String() != tt.want {
				t.Errorf("clientIP = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// A proxy may append its own X-Forwarded-For header instead of extending the
// client's, so the repeats are read as one chain in the order they arrived.
func TestClientIPJoinsRepeatedForwardedForHeaders(t *testing.T) {
	ctx := ctxWith("10.0.0.1:5000", "1.2.3.4, 203.0.113.9", "172.16.0.1")

	got, err := clientIP(ctx, 2)
	if err != nil {
		t.Fatalf("clientIP returned error: %v", err)
	}

	if got.String() != "203.0.113.9" {
		t.Errorf("clientIP = %q, want 203.0.113.9", got.String())
	}
}

func TestClientIPMatchesHeaderNameCaseInsensitively(t *testing.T) {
	ctx := fakeContext{
		remoteAddr: "10.0.0.1:5000",
		headers:    []header{{"x-forwarded-for", "203.0.113.9"}},
	}

	got, err := clientIP(ctx, 1)
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
