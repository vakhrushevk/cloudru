package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/vakhrushevk/cloudru/internal/app"
	"github.com/vakhrushevk/cloudru/internal/config"
)

func main() {
	app, err := app.NewApp(context.Background(), "configs/config.yaml")

	if err != nil {
		log.Fatal("error creating app:", err)
	}

	log.Fatal(app.Start())
}

func exampleBackends(cfg config.BalancerConfig) {
	for i := 0; i < len(cfg.Backends); i++ {
		m := http.NewServeMux()
		m.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!" + fmt.Sprintf("from backend %d", i)))
		}))
		backendURL, err := url.Parse(cfg.Backends[i].URL)
		if err != nil {
			log.Printf("Failed to parse backend URL %s: %v", cfg.Backends[i].URL, err)
			continue
		}
		slog.Debug("Starting backend", "backend", cfg.Backends[i].URL)
		go http.ListenAndServe(backendURL.Host, m)
	}
	time.Sleep(2 * time.Second)
	m := http.NewServeMux()
	m.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!" + fmt.Sprintf("from backend %d", 3)))
	}))
	go http.ListenAndServe("localhost:8003", m)
}
