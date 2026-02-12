package resilience

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"
)

// RetryConfig holds retry configuration.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns sensible retry defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}
}

// WithRetry executes a function with exponential backoff retries.
func WithRetry(ctx context.Context, config RetryConfig, logger *zap.Logger, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt == config.MaxRetries {
				break
			}

			delay := time.Duration(float64(config.BaseDelay) * math.Pow(2, float64(attempt)))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			logger.Warn("operation failed, retrying",
				zap.String("operation", operation),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
				zap.Error(err),
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		} else {
			return nil
		}
	}

	return lastErr
}
