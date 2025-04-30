package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vakhrushevk/cloudru/internal/config"
)

func main() {

	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("error loading config:", err)
	}
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("Config loaded:", config)
		fmt.Println("____________________________________________________________")
		fmt.Println("HTTP Port:", config.HttpConfig.ListenPort)
		fmt.Println("____________________________________________________________")
		fmt.Println("Retry Config:")
		fmt.Println("Retry Max Attempts:", config.RetryConfig.MaxAttempts)
		fmt.Println("Retry Delay:", config.RetryConfig.Delay)
		fmt.Println("Retry Max Delay:", config.RetryConfig.MaxDelay)
		fmt.Println("____________________________________________________________")
		fmt.Println("Balancer Config:")
		fmt.Println("Balancer Strategy:", config.BalancerConfig.Strategy)
		fmt.Println("Balancer Backends File:", config.BalancerConfig.BackedsFile)
		for _, backend := range config.BalancerConfig.Backends {
			fmt.Println("Backend URL:", backend.Url)
		}
	}

}
