package ratelimit

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type Bucket struct {
	Capacity   int
	Tokens     int
	RefilRate  int
	LastRefill time.Time
	Mutex      sync.Mutex
}

func (b *Bucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.LastRefill).Seconds()
	added := int(elapsed * float64(b.RefilRate))
	if added > 0 {
		b.Tokens = min(b.Capacity, b.Tokens+added)
		b.LastRefill = now
	}
}

type Limiter struct {
	mu       sync.RWMutex
	buckets  map[string]*Bucket
	capacity int
	rate     int
}

func NewLimiter() *Limiter {
	return &Limiter{
		buckets:  make(map[string]*Bucket),
		capacity: 10,
		rate:     1,
	}
}

func (l *Limiter) getBucket(clientIP string) *Bucket {
	l.mu.RLock()
	b, ok := l.buckets[clientIP]
	l.mu.RUnlock()
	if ok {
		return b
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if b, ok = l.buckets[clientIP]; ok {
		return b
	}

	b = &Bucket{
		Tokens:     l.capacity,
		LastRefill: time.Now(),
		Capacity:   l.capacity,
		RefilRate:  l.rate,
	}
	l.buckets[clientIP] = b

	return b
}

func (l *Limiter) Allow(clientIP string) bool {
	b := l.getBucket(clientIP)
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	b.refill()

	if b.Tokens > 0 {
		b.Tokens--
		return true
	}

	return false
}

func RateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("Rate limit middleware", "remote_addr", r.RemoteAddr)
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"code":  "429",
					"error": "Rate limit exceeded",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
