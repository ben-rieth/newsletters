package auth

import (
	"net/netip"
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

// Aliased so the embedded field is not named Context, which would shadow the
// interface's own Context() method.
type embeddedHumaContext = huma.Context

type fakeContext struct {
	embeddedHumaContext
	remoteAddr string
	headers    map[string]string
}

func (c fakeContext) RemoteAddr() string { return c.remoteAddr }

func (c fakeContext) EachHeader(cb func(name, value string)) {
	for name, value := range c.headers {
		cb(name, value)
	}
}

func ctxWith(remoteAddr, forwardedFor string) fakeContext {
	headers := map[string]string{}
	if forwardedFor != "" {
		headers["X-Forwarded-For"] = forwardedFor
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
	limiter := IPRateLimiter{
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
