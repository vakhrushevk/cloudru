package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/vakhrushevk/cloudru/internal/config"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	backends := cfg.BalancerConfig.Backends
	var wg sync.WaitGroup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	servers := make([]*http.Server, len(backends))

	for i, backend := range backends {
		srv := startBackend(backend, i)
		servers[i] = srv
		wg.Add(1)
		go func(s *http.Server) {
			defer wg.Done()
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("ListenAndServe error: %v", err)
			}
		}(srv)
	}

	<-stop
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}

	wg.Wait()
	log.Println("All servers stopped.")
}

func startBackend(backend config.BackendConfig, i int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, you are on backend " + backend.URL))
	})
	addr := backend.URL
	addr = strings.TrimPrefix(addr, "http://")
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	slog.Info("Backend started", "url", backend.URL, "index", i)
	return server
}
