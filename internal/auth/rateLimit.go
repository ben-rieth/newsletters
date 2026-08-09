package auth

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/time/rate"
)

type Limiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*Limiter
	limit    rate.Limit
	burst    int
}

func (i *IPRateLimiter) GetLimiter(ip string) *Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	if limiter, exists := i.limiters[ip]; exists {
		limiter.lastSeen = time.Now()
		return limiter
	}

	limiter := &Limiter{
		limiter:  rate.NewLimiter(i.limit, i.burst),
		lastSeen: time.Now(),
	}

	i.limiters[ip] = limiter
	return limiter
}

func (i *IPRateLimiter) cleanUp() {
	i.mu.Lock()
	defer i.mu.Unlock()

	oneHourAgo := time.Now().Add(time.Hour * -1)

	for ip, limiter := range i.limiters {
		if limiter.lastSeen.Before(oneHourAgo) {
			delete(i.limiters, ip)
		}
	}
}

// peerIP is the address of whoever actually opened the connection, so it can
// never be forged by the client.
func peerIP(ctx huma.Context) (string, error) {
	ip, _, err := net.SplitHostPort(ctx.RemoteAddr())
	if err != nil {
		return "", err
	}

	return ip, nil
}

func forwardedForChain(ctx huma.Context) []string {
	var chain []string

	ctx.EachHeader(func(name, value string) {
		if !strings.EqualFold(name, "X-Forwarded-For") {
			return
		}

		for _, hop := range strings.Split(value, ",") {
			if hop = strings.TrimSpace(hop); hop != "" {
				chain = append(chain, hop)
			}
		}
	})

	return chain
}

// clientIP walks back exactly trustedProxyCount hops from the server. Anything
// further left in X-Forwarded-For was appended by an untrusted party and is
// attacker-controlled, so a short or malformed chain falls back to the peer
// address rather than trusting what the client sent.
func clientIP(ctx huma.Context, trustedProxyCount int) (string, error) {
	peer, err := peerIP(ctx)
	if err != nil {
		return "", err
	}

	if trustedProxyCount == 0 {
		return peer, nil
	}

	chain := forwardedForChain(ctx)
	i := len(chain) - trustedProxyCount
	if i < 0 {
		return peer, nil
	}

	if addr, err := netip.ParseAddr(chain[i]); err == nil {
		return addr.String(), nil
	}

	return peer, nil
}

func NewRateLimitMiddleware(ctx context.Context, api huma.API, cfg *config.Config, limit, burst int) func(ctx huma.Context, next func(huma.Context)) {
	limiters := make(map[string]*Limiter)

	ipRateLimiter := IPRateLimiter{
		limiters: limiters,
		limit:    rate.Limit(limit),
		burst:    burst,
	}

	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				ipRateLimiter.cleanUp()
			}
		}
	}()

	return func(ctx huma.Context, next func(huma.Context)) {
		ip, err := clientIP(ctx, cfg.TrustedProxyCount)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Something went wrong.")
			return
		}

		limiter := ipRateLimiter.GetLimiter(ip)

		if limiter.limiter.Allow() {
			next(ctx)
			return
		}

		huma.WriteErr(api, ctx, http.StatusTooManyRequests, "Too many requests. Please wait and then try again.")
	}
}
