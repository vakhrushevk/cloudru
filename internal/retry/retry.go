package retry

import (
	"log"
	"time"

	"github.com/vakhrushevk/cloudru/internal/config"
)

func WithRetry(config config.RetryConfig, fn func() error) error {
	var err error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		log.Printf("Attempt %d of %d", attempt, config.MaxAttempts)
		err = fn()
		if err == nil {
			return nil
		}

		if attempt == config.MaxAttempts {
			return err
		}

		delay := config.Delay * time.Duration(attempt)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		log.Printf("Waiting %v before next attempt", delay)
		time.Sleep(delay)
	}

	return err
}
