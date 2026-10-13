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

// One IPv6 allocation holds more addresses than the map ever could, so v6
// clients are bucketed by routed prefix rather than by address.
const ipv6BucketBits = 64

// Past the ceiling a bucket is evicted rather than the new client refused, since
// refusing would make the ceiling itself a way to lock everyone else out.
const maxLimiters = 10_000

type Limiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type KeyedLimiter struct {
	mu       sync.Mutex
	limiters map[string]*Limiter
	limit    rate.Limit
	burst    int
}

func (i *KeyedLimiter) Allow(key string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.limiterFor(key).limiter.Allow()
}

// A key with no bucket yet would get a full one, so checking it doesn't create
// one; a check alone shouldn't let a caller grow the map.
func (i *KeyedLimiter) WouldAllow(key string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[key]
	return !exists || limiter.limiter.Tokens() >= 1
}

func (i *KeyedLimiter) Reset(key string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.limiters, key)
}

func (i *KeyedLimiter) GetLimiter(key string) *Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.limiterFor(key)
}

// Callers must hold i.mu.
func (i *KeyedLimiter) limiterFor(key string) *Limiter {
	if limiter, exists := i.limiters[key]; exists {
		limiter.lastSeen = time.Now()
		return limiter
	}

	if len(i.limiters) >= maxLimiters {
		i.evictSampled()
	}

	limiter := &Limiter{
		limiter:  rate.NewLimiter(i.limit, i.burst),
		lastSeen: time.Now(),
	}

	i.limiters[key] = limiter
	return limiter
}

// Whoever fills the map is cycling keys faster than they expire, so eviction
// runs on their every request. Scanning it whole would hold the mutex for all of
// it and slow down everyone else, so the victim is the oldest of a small sample.
// Go randomises map iteration order, which is what makes the sample worth taking.
const evictionSampleSize = 8

// Callers must hold i.mu.
func (i *KeyedLimiter) evictSampled() {
	var oldestKey string
	var oldestSeen time.Time

	sampled := 0
	for key, limiter := range i.limiters {
		if oldestKey == "" || limiter.lastSeen.Before(oldestSeen) {
			oldestKey, oldestSeen = key, limiter.lastSeen
		}

		if sampled++; sampled >= evictionSampleSize {
			break
		}
	}

	if oldestKey != "" {
		delete(i.limiters, oldestKey)
	}
}

// A bucket that has refilled is indistinguishable from the fresh one GetLimiter
// would create, so only those are dropped. Dropping on idle time instead would
// hand back a full budget early to any limiter that refills slower than that.
func (i *KeyedLimiter) cleanUp() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for key, limiter := range i.limiters {
		if limiter.limiter.Tokens() >= float64(i.burst) {
			delete(i.limiters, key)
		}
	}
}

// peerIP is the address of whoever actually opened the connection, so it can
// never be forged by the client.
func peerIP(ctx huma.Context) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(ctx.RemoteAddr())
	if err != nil {
		return netip.Addr{}, err
	}

	return netip.ParseAddr(host)
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
func clientIP(ctx huma.Context, trustedProxyCount int) (netip.Addr, error) {
	peer, err := peerIP(ctx)
	if err != nil {
		return netip.Addr{}, err
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
		return addr.Unmap(), nil
	}

	return peer, nil
}

func rateLimitKey(addr netip.Addr) string {
	addr = addr.Unmap()

	if addr.Is4() {
		return addr.String()
	}

	prefix, err := addr.Prefix(ipv6BucketBits)
	if err != nil {
		return addr.String()
	}

	return prefix.String()
}

func NewKeyedLimiter(ctx context.Context, limit rate.Limit, burst int) *KeyedLimiter {
	keyedLimiter := &KeyedLimiter{
		limiters: make(map[string]*Limiter),
		limit:    limit,
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
				keyedLimiter.cleanUp()
			}
		}
	}()

	return keyedLimiter
}

func NewRateLimitMiddleware(ctx context.Context, api huma.API, cfg *config.Config, limit, burst int) func(ctx huma.Context, next func(huma.Context)) {
	ipRateLimiter := NewKeyedLimiter(ctx, rate.Limit(limit), burst)

	return func(ctx huma.Context, next func(huma.Context)) {
		addr, err := clientIP(ctx, cfg.TrustedProxyCount)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Something went wrong.")
			return
		}

		clientKey := rateLimitKey(addr)

		if ipRateLimiter.Allow(clientKey) {
			next(huma.WithValue(ctx, ClientKeyKey, clientKey))
			return
		}

		huma.WriteErr(api, ctx, http.StatusTooManyRequests, "Too many requests. Please wait and then try again.")
	}
}
