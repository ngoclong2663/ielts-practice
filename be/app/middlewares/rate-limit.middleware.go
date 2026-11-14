package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"ielts-practice/pkg/exceptions"
)

// Per-client-IP token buckets. Each preset keeps its own bucket per IP.
// Note: limits are per API instance (in memory). Move to Redis if we scale horizontally.
var (
	DEFAULT_RATE_LIMIT     rate.Limit = 20 // requests per second per IP
	DEFAULT_BURST_SIZE     int        = 40
	SHORT_TERM_RATE_LIMIT  rate.Limit = 5
	SHORT_TERM_BURST_SIZE  int        = 10
	MEDIUM_TERM_RATE_LIMIT rate.Limit = 1
	MEDIUM_TERM_BURST_SIZE int        = 5
	LONG_TERM_RATE_LIMIT   rate.Limit = 50
	LONG_TERM_BURST_SIZE   int        = 100

	limiterIdleTTL = 10 * time.Minute
)

// rateLimitEnabled can be switched off for automated tests (never in production).
var rateLimitEnabled = true

func SetRateLimitEnabled(enabled bool) { rateLimitEnabled = enabled }

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
}

func newIPRateLimiter(limit rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{visitors: map[string]*visitor{}, limit: limit, burst: burst}
	go l.cleanup()
	return l
}

func (l *ipRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (l *ipRateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > limiterIdleTTL {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *ipRateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rateLimitEnabled && !l.get(c.ClientIP()).Allow() {
			WriteError(c, exceptions.NewException(http.StatusTooManyRequests, exceptions.CodeRateLimited, "Too many requests", nil))
			return
		}
		c.Next()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return newIPRateLimiter(DEFAULT_RATE_LIMIT, DEFAULT_BURST_SIZE).middleware()
}

func ShortTermRateLimitMiddleware() gin.HandlerFunc {
	return newIPRateLimiter(SHORT_TERM_RATE_LIMIT, SHORT_TERM_BURST_SIZE).middleware()
}

func MediumTermRateLimitMiddleware() gin.HandlerFunc {
	return newIPRateLimiter(MEDIUM_TERM_RATE_LIMIT, MEDIUM_TERM_BURST_SIZE).middleware()
}

func LongTermRateLimitMiddleware() gin.HandlerFunc {
	return newIPRateLimiter(LONG_TERM_RATE_LIMIT, LONG_TERM_BURST_SIZE).middleware()
}
