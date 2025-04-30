package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/vakhrushevk/cloudru/internal/balancer"
	"github.com/vakhrushevk/cloudru/internal/config"
	"github.com/vakhrushevk/cloudru/pkg/logger"
)

func main() {

	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("error loading config:", err)
	}
	logger.Init(&config.LoggerConfig)

	b, err := balancer.New(config.BalancerConfig, config.RetryConfig)
	if err != nil {
		log.Fatalf("failed to create balancer: %v", err)
	}

	balancer.CheckAndUpdate(*config, b)

	go func() {
		time.Sleep(2 * time.Second)
		// balancer.RemoveAllBackend()
		// fmt.Println("All backends removed")
		// time.Sleep(10 * time.Second)
		// fmt.Println("Adding  backends")
		// balancer.RegisterBackend("http://localhost:8001")
		// balancer.RegisterBackend("http://localhost:8002")
		// balancer.RegisterBackend("http://localhost:8003")
		// balancer.RegisterBackend("http://localhost:8004")
		// fmt.Println("All backends added")
	}()

	go exampleBackends(config.BalancerConfig)

	http.HandleFunc("/", b.BalanceHandler().ServeHTTP)
	log.Println("Starting server on port", config.HTTPConfig.ListenPort)
	log.Fatal(http.ListenAndServe(":"+config.HTTPConfig.ListenPort, nil))
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
