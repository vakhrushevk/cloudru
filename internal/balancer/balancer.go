package balancer

import (
	"errors"
	"net/http"

	roundrobin "github.com/vakhrushevk/cloudru/internal/balancer/roundRobin"
	"github.com/vakhrushevk/cloudru/internal/config"
)

var (
	ErrBalancerStrategyNotFound = errors.New("balancer strategy not found")
)

type Balancer interface {
	BalanceHandler() http.Handler
	RemoveAllBackend()
	RegisterBackend(URL string)
}

func New(cfg config.BalancerConfig, retryConfig config.RetryConfig) (Balancer, error) {
	switch cfg.Strategy {
	case "round_robin":
		return roundrobin.New(cfg, retryConfig)
	case "random":
		panic("not implemented")
	default:
		return nil, ErrBalancerStrategyNotFound
	}
}

func CheckAndUpdate(cfg config.Config, balancer Balancer) {
	Watcher, err := config.NewWatcher(cfg.BackedsFile)
	Watcher.DoRun(func() {
		cfg.Backends, err = config.LoadBackends(cfg.BackedsFile)
		if err != nil {
			return
		}
		balancer.RemoveAllBackend()
		for _, b := range cfg.Backends {
			balancer.RegisterBackend(b.Url)
		}
	})
}
