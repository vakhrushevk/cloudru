// Package ratelimit предоставляет функцию для ограничения количества запросов
package ratelimit

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/vakhrushevk/cloudru/internal/config"
	"github.com/vakhrushevk/cloudru/internal/repository"
)

// Limiter структура для ограничения количества запросов
type Limiter struct {
	bucketRepo   repository.BucketRepository
	bucketConfig config.BucketConfig
}

// NewLimiter создает новый лимитер
func NewLimiter(ctx context.Context, bucketRepo repository.BucketRepository, bucketConfig config.BucketConfig) *Limiter {
	limiter := &Limiter{
		bucketRepo:   bucketRepo,
		bucketConfig: bucketConfig,
	}

	limiter.StartRefillBuckets(ctx)

	return limiter
}

// StartRefillBuckets заполняет бакеты токенами
func (l *Limiter) StartRefillBuckets(ctx context.Context) {
	ticker := time.NewTicker(l.bucketConfig.RefilTime)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("Stopping token refill")
				return
			case <-ticker.C:
				slog.Debug("Refilling tokens for all buckets")
				err := l.bucketRepo.RefillAllBuckets(ctx)
				if err != nil {
					slog.Error("Failed to refill buckets", "error", err)
				}
			}
		}
	}()
}

// Allow проверяет, может ли клиент выполнить запрос
func (l *Limiter) Allow(ctx context.Context, clientIP string) bool {
	b, err := l.bucketRepo.Bucket(ctx, clientIP)
	if err == nil {
		if b.Tokens > 0 {
			ok, err := l.bucketRepo.Decrease(ctx, clientIP)
			if err != nil || !ok {
				return false
			}
			return true
		}
		return false
	}

	if err == repository.ErrBucketNotFound {
		err := l.bucketRepo.CreateBucket(ctx, clientIP, l.bucketConfig.Capacity, l.bucketConfig.RefilRate, l.bucketConfig.Tokens)
		if err != nil {
			slog.Error("Failed to create bucket", "error", err)
			return false
		}
		ok, err := l.bucketRepo.Decrease(ctx, clientIP)
		if err != nil || !ok {
			return false
		}
		return true
	}

	slog.Error("Failed to get bucket", "error", err)
	return false

}

// Middleware middleware для ограничения количества запросов
func Middleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("Rate limit middleware", "remote_addr", r.RemoteAddr)

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			slog.Debug("Rate limit middleware", "ip", ip)
			if !limiter.Allow(context.TODO(), ip) {
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
