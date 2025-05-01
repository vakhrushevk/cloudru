package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	ratelimit "github.com/vakhrushevk/cloudru/internal/rateLimit"
)

var globalConfigPath string

// App структура приложения
type App struct {
	serviceProvider *serviceProvider
	httpServer      *http.Server
}

// NewApp создает новый App
func NewApp(ctx context.Context, configPath string) (*App, error) {
	a := &App{}
	globalConfigPath = configPath
	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// initServiceProvider инициализирует serviceProvider
func (a *App) initServiceProvider(ctx context.Context) error {
	serviceProvider, err := newServiceProvider(ctx)
	if err != nil {
		return err
	}
	a.serviceProvider = serviceProvider
	return nil
}

// initHttpServer инициализирует http сервер
func (a *App) initHttpServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/", ratelimit.Middleware(
		a.serviceProvider.Limiter(ctx))(
		a.serviceProvider.Balancer(ctx).BalanceHandler()))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.serviceProvider.Config().HTTPConfig.ListenPort),
		Handler: mux,
	}
	a.httpServer = server

	return nil
}

// initDeps инициализирует зависимости
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

// Start запускает сервер
func (a *App) Start() error {
	slog.Info("Starting server on port", "port", a.serviceProvider.Config().HTTPConfig.ListenPort)
	return a.httpServer.ListenAndServe()
}
