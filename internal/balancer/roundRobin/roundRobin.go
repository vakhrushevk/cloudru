//nolint:revive
package roundrobin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vakhrushevk/cloudru/internal/balancer/backend"
	"github.com/vakhrushevk/cloudru/internal/config"
	"github.com/vakhrushevk/cloudru/internal/retry"
)

// Balancer балансировщик на основе round robin
type Balancer struct {
	backends    []*backend.Backend
	current     uint64
	mu          sync.RWMutex
	retryConfig config.RetryConfig
}

// New создает новый Balancer
func New(balanceCofnig config.BalancerConfig, retryConfig config.RetryConfig) (*Balancer, error) {
	rb := &Balancer{}
	rb.retryConfig = retryConfig
	rb.backends = make([]*backend.Backend, 0, len(balanceCofnig.Backends))

	for _, b := range balanceCofnig.Backends {
		rb.RegisterBackend(b.URL)
	}
	rb.current = 0
	rb.mu = sync.RWMutex{}
	go rb.healthCheck(context.TODO(), balanceCofnig.HealthCheckInterval) // TODO: add to config
	return rb, nil
}

// RegisterBackend регистрирует новый бэкенд
func (rb *Balancer) RegisterBackend(URL string) {
	u, err := url.Parse(URL)
	if err != nil {
		// TODO: ADD LOGGER
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	backend := backend.NewBackend(u, true, proxy)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		rb.BalancerErrorHandler(w, err, backend)
	}
	rb.backends = append(rb.backends, backend)
}

// nextIndex возвращает следующий индекс в списке backends
func (rb *Balancer) nextIndex() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if len(rb.backends) == 0 {
		return -1
	}
	return int(atomic.AddUint64(&rb.current, uint64(1)) % uint64(len(rb.backends)))
}

// NextPeer возвращает следующий доступный backend
func (rb *Balancer) nextPeer() *backend.Backend {
	next := rb.nextIndex()
	for i := next; i < len(rb.backends)+next; i++ {
		idx := i % len(rb.backends)
		if rb.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&rb.current, uint64(idx))
			}
			return rb.backends[idx]
		}
	}
	return nil
}

// BalanceHandler обрабатывает запросы и перенаправляет их на следующий доступный backend
func (rb *Balancer) BalanceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rb.mu.RLock()
		if len(rb.backends) == 0 {
			http.Error(w, "No backends available", http.StatusServiceUnavailable)
			return
		}
		rb.mu.RUnlock()

		peer := rb.nextPeer()
		if peer != nil {
			peer.ReverseProxy.ServeHTTP(w, r)
			return
		}
		slog.Warn("All backends are unavailable")
		http.Error(w, "All backends are unavailable", http.StatusServiceUnavailable)
	})
}

// BalancerErrorHandler обрабатывает ошибки при перенаправлении запросов на backend
func (rb *Balancer) BalancerErrorHandler(w http.ResponseWriter, err error, backend *backend.Backend) {
	slog.Error("Error redirecting request to backend", "backend", backend.URL.String(), "error", err)
	backend.SetAlive(false)
	slog.Info("Attempting to restore connection to backend", "backend", backend.URL.String())
	retryErr := retry.WithRetry(rb.retryConfig, backend.IsBackendAlive)
	if retryErr != nil {
		slog.Error("Failed to restore connection to backend", "backend", backend.URL.String(), "error", retryErr)
		http.Error(w, fmt.Sprintf("Backend %s is unavailable", backend.URL.Host), http.StatusServiceUnavailable)
		return
	}
	slog.Info("Connection to backend restored", "backend", backend.URL.String())
	backend.SetAlive(true)
}

// healthCheck проверяет состояние бэкендов
func (rb *Balancer) healthCheck(ctx context.Context, delay time.Duration) {
	t := time.NewTicker(delay)
	for {
		select {
		case <-t.C:
			slog.Debug("Starting health check...")
			for _, backend := range rb.backends {
				err := backend.IsBackendAlive()
				if err != nil {
					slog.Error("Backend is unavailable", "backend", backend.URL, "error", err)
					backend.SetAlive(false)
				} else {
					backend.SetAlive(true)
				}
			}
			slog.Debug("Health check completed")
		case <-ctx.Done():
			return
		}
	}
}

// RemoveAllBackend удаляет все бэкенды
func (rb *Balancer) RemoveAllBackend() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.backends = make([]*backend.Backend, 0)
	atomic.StoreUint64(&rb.current, 0)
}
