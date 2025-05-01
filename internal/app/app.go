package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	ratelimit "github.com/vakhrushevk/cloudru/internal/rateLimit"
)

var globalConfigPath string

type App struct {
	serviceProvider *serviceProvider
	httpServer      *http.Server
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	a := &App{}
	globalConfigPath = configPath
	a.initDeps(ctx)

	return a, nil
}

func (a *App) initServiceProvider(ctx context.Context) error {
	serviceProvider, err := NewServiceProvider(ctx)
	if err != nil {
		return err
	}
	a.serviceProvider = serviceProvider
	return nil
}

func (a *App) initHttpServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/rate-limit", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	mux.Handle("/", ratelimit.RateLimitMiddleware(
		a.serviceProvider.Limiter(ctx))(
		a.serviceProvider.Balancer(ctx).BalanceHandler()))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.serviceProvider.Config().HTTPConfig.ListenPort),
		Handler: mux,
	}
	a.httpServer = server

	return nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initServiceProvider,
		a.initHttpServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Start() error {
	slog.Info("Starting server on port", "port", a.serviceProvider.Config().HTTPConfig.ListenPort)
	return a.httpServer.ListenAndServe()
}

// redisClient := redis.NewClient(&redis.Options{
// 	Addr:     fmt.Sprintf("%s:%d", config.RedisConfig.Host, config.RedisConfig.Port),
// 	Password: config.RedisConfig.Password,
// 	DB:       config.RedisConfig.DB,
// })
// redisRepo, err := redisRepository.NewRedisRepository(redisClient)
// if err != nil {
// 	log.Fatal("error creating redis repository:", err)
// }
// limiter := ratelimit.NewLimiter(redisRepo, config.BucketConfig)

// b, err := balancer.New(config.BalancerConfig, config.RetryConfig)
// if err != nil {
// 	log.Fatalf("failed to create balancer: %v", err)
// }

// balancer.CheckAndUpdate(*config, b)

// go func() {
// 	time.Sleep(2 * time.Second)
// 	// balancer.RemoveAllBackend()
// 	// fmt.Println("All backends removed")
// 	// time.Sleep(10 * time.Second)
// 	// fmt.Println("Adding  backends")
// 	// balancer.RegisterBackend("http://localhost:8001")
// 	// balancer.RegisterBackend("http://localhost:8002")
// 	// balancer.RegisterBackend("http://localhost:8003")
// 	// balancer.RegisterBackend("http://localhost:8004")
// 	// fmt.Println("All backends added")
// }()

// go exampleBackends(config.BalancerConfig)

// r := http.NewServeMux()
// r.HandleFunc("/api/v1/rate-limit", func(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Hello, World!"))
// })
// r.Handle("/", ratelimit.RateLimitMiddleware(limiter)(b.BalanceHandler()))
// log.Println("Starting server on port", config.HTTPConfig.ListenPort)
// log.Fatal(http.ListenAndServe(":"+config.HTTPConfig.ListenPort, r))
