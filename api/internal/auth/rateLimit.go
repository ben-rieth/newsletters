package auth

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

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

func NewRateLimitMiddleware(ctx context.Context, api huma.API, limit, burst int) func(ctx huma.Context, next func(huma.Context)) {
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
		address := ctx.RemoteAddr()
		ip, _, err := net.SplitHostPort(address)
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
