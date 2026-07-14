package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func clientIP(r *http.Request) string {
	// Если приложение находится за reverse proxy
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	// Cloudflare
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Nginx
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.RWMutex
	users    map[string]*userLimiter
	rate     rate.Limit
	burst    int
	lifetime time.Duration
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		users:    make(map[string]*userLimiter),
		rate:     rate.Every(time.Minute / 100), // 100 запросов в минуту
		burst:    100,
		lifetime: 10 * time.Minute,
	}

	go rl.cleanup()

	return rl
}

func (r *RateLimiter) getLimiter(ip string) *rate.Limiter {
	r.mu.RLock()
	ul, ok := r.users[ip]
	r.mu.RUnlock()

	if ok {
		r.mu.Lock()
		ul.lastSeen = time.Now()
		r.mu.Unlock()
		return ul.limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if ul, ok := r.users[ip]; ok {
		ul.lastSeen = time.Now()
		return ul.limiter
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/100), 100)

	r.users[ip] = &userLimiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ip := clientIP(req)

		limiter := r.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()

		for id, limiter := range r.users {
			if time.Since(limiter.lastSeen) > r.lifetime {
				delete(r.users, id)
			}
		}

		r.mu.Unlock()
	}
}
