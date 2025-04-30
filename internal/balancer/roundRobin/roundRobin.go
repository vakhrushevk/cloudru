package roundrobin

import (
	"context"
	"fmt"
	"log"
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

type RoundRobinBalancer struct {
	backends    []*backend.Backend
	current     uint64
	mu          sync.RWMutex
	retryConfig config.RetryConfig
}

func New(balanceCofnig config.BalancerConfig, retryConfig config.RetryConfig) (*RoundRobinBalancer, error) {
	rb := &RoundRobinBalancer{}
	rb.retryConfig = retryConfig
	rb.backends = make([]*backend.Backend, 0, len(balanceCofnig.Backends))

	for _, b := range balanceCofnig.Backends {
		rb.RegisterBackend(b.Url)
	}
	rb.current = 0
	rb.mu = sync.RWMutex{}
	go rb.healthCheck(context.TODO(), 20*time.Second) // TODO: add to config
	return rb, nil
}

func (b *RoundRobinBalancer) RegisterBackend(URL string) {
	u, err := url.Parse(URL)
	if err != nil {
		// TODO: ADD LOGGER
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	backend := backend.NewBackend(u, true, proxy)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		b.BalancerErrorHandler(w, r, err, backend)
	}
	b.backends = append(b.backends, backend)
}

// nextIndex возвращает следующий индекс в списке backends
func (r *RoundRobinBalancer) nextIndex() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.backends) == 0 {
		return -1
	}
	return int(atomic.AddUint64(&r.current, uint64(1)) % uint64(len(r.backends)))
}

// NextPeer возвращает следующий доступный backend
func (b *RoundRobinBalancer) nextPeer() *backend.Backend {
	next := b.nextIndex()
	for i := next; i < len(b.backends)+next; i++ {
		idx := i % len(b.backends)
		if b.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&b.current, uint64(idx))
			}
			return b.backends[idx]
		}
	}
	return nil
}

// BalanceHandler обрабатывает запросы и перенаправляет их на следующий доступный backend
func (b *RoundRobinBalancer) BalanceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.mu.RLock()
		if len(b.backends) == 0 {
			http.Error(w, "No backends available", http.StatusServiceUnavailable)
			return
		}
		b.mu.RUnlock()

		peer := b.nextPeer()
		if peer != nil {
			peer.ReverseProxy.ServeHTTP(w, r)
			return
		}
		log.Printf("No available backends")
		http.Error(w, "All backends are unavailable", http.StatusServiceUnavailable)
	})
}

// BalancerErrorHandler обрабатывает ошибки при перенаправлении запросов на backend
func (b *RoundRobinBalancer) BalancerErrorHandler(w http.ResponseWriter, r *http.Request, err error, backend *backend.Backend) {
	log.Printf("Error redirecting request to backend %s: %v", backend.URL.String(), err)
	backend.SetAlive(false)

	log.Printf("Attempting to restore connection to backend %s", backend.URL.String())
	retryErr := retry.WithRetry(b.retryConfig, backend.IsBackendAlive)
	if retryErr != nil {
		log.Printf("Failed to restore connection to backend %s after retries: %v", backend.URL.String(), retryErr)
		http.Error(w, fmt.Sprintf("Backend %s is unavailable", backend.URL.Host), http.StatusServiceUnavailable)
		return
	}

	log.Printf("Connection to backend %s restored", backend.URL.String())
	backend.SetAlive(true)
}

// healthCheck проверяет состояние бэкендов
func (b *RoundRobinBalancer) healthCheck(ctx context.Context, delay time.Duration) {
	t := time.NewTicker(delay)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			for _, backend := range b.backends {
				err := backend.IsBackendAlive()
				if err != nil {
					backend.SetAlive(false)
				} else {
					backend.SetAlive(true)
				}
				log.Printf("Backend %s [%v]\n", backend.URL, backend.IsAlive())
			}
			log.Println("Health check completed")
		case <-ctx.Done():
			return
		}
	}
}
func (b *RoundRobinBalancer) RemoveAllBackend() {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Println("удаляем все серверры бекуенда")
	b.backends = make([]*backend.Backend, 0)
	atomic.StoreUint64(&b.current, 0)
}
