//nolint:revive
package balancer

import (
	"context"
	"errors"
	"net/http"

	roundrobin "github.com/vakhrushevk/cloudru/internal/balancer/roundRobin"
	"github.com/vakhrushevk/cloudru/internal/config"
)

var (
	// ErrBalancerStrategyNotFound ошибка балансировщика, если стратегия не найдена
	ErrBalancerStrategyNotFound = errors.New("balancer strategy not found")
)

// Balancer интерфейс балансировщика
type Balancer interface {
	BalanceHandler() http.Handler
	RemoveAllBackend()
	RegisterBackend(URL string)
}

// New создает новый балансировщик
func New(ctx context.Context, cfg config.BalancerConfig, retryConfig config.RetryConfig) (Balancer, error) {
	switch cfg.Strategy {
	case "round_robin":
		return roundrobin.New(ctx, cfg, retryConfig)
	case "random":
		panic("not implemented")
	default:
		return nil, ErrBalancerStrategyNotFound
	}
}

// CheckAndUpdate проверяет и обновляет бэкенды
func CheckAndUpdate(cfg config.Config, balancer Balancer) {
	Watcher, err := config.NewWatcher(cfg.BalancerConfig.BackedsFile)
	Watcher.DoRun(func() {
		cfg.BalancerConfig.Backends, err = config.LoadBackends(cfg.BalancerConfig.BackedsFile)
		if err != nil {
			return
		}
		balancer.RemoveAllBackend()
		for _, b := range cfg.BalancerConfig.Backends {
			balancer.RegisterBackend(b.URL)
		}
	})
}
