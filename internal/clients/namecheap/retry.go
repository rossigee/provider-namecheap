package namecheap

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// RetryConfig defines retry behavior for API calls
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
	RetryableErrors []error
}

// DefaultRetryConfig returns sensible defaults for production use
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
		RetryableErrors: []error{
			&net.DNSError{},
			&net.OpError{},
		},
	}
}

// RetryableFunc represents a function that can be retried
type RetryableFunc func(ctx context.Context) error

// WithRetry executes a function with exponential backoff retry logic
func (c *Client) WithRetry(ctx context.Context, operation string, fn RetryableFunc) error {
	config := c.retryConfig
	if config == nil {
		defaultConfig := DefaultRetryConfig()
		config = &defaultConfig
	}

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Create a new context with timeout for each attempt
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		err := fn(attemptCtx)
		cancel()

		if err == nil {
			if attempt > 0 {
				c.logRetrySuccess(operation, attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryableError(err) {
			return errors.Wrapf(err, "non-retryable error in %s", operation)
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxRetries {
			delay := c.calculateDelay(config, attempt)
			c.logRetryAttempt(operation, attempt+1, delay, err)

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return errors.Wrapf(lastErr, "operation %s failed after %d retries", operation, config.MaxRetries)
}

// isRetryableError determines if an error should trigger a retry
func (c *Client) isRetryableError(err error) bool {
	// Network errors are generally retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// HTTP client timeout errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	// HTTP status codes that are retryable
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusTooManyRequests,
			 http.StatusInternalServerError,
			 http.StatusBadGateway,
			 http.StatusServiceUnavailable,
			 http.StatusGatewayTimeout:
			return true
		}
	}

	// Namecheap-specific retryable errors
	var ncErr Error
	if errors.As(err, &ncErr) {
		switch ncErr.Number {
		case "2030280", "2030281": // Rate limiting errors
			return true
		case "2011170": // Server temporarily unavailable
			return true
		}
	}

	return false
}

// calculateDelay computes the delay before the next retry attempt
func (c *Client) calculateDelay(config *RetryConfig, attempt int) time.Duration {
	// Exponential backoff
	delay := float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt))

	// Add jitter to prevent thundering herd
	if config.JitterFactor > 0 {
		jitter := delay * config.JitterFactor * (rand.Float64()*2 - 1)
		delay += jitter
	}

	// Cap at maximum delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// logRetryAttempt logs retry attempts for observability
func (c *Client) logRetryAttempt(operation string, attempt int, delay time.Duration, err error) {
	if c.logger.GetSink() != nil {
		c.logger.Info("Retrying API operation",
			"operation", operation,
			"attempt", attempt,
			"delay", delay,
			"error", err.Error())
	}
}

// logRetrySuccess logs successful retry for observability
func (c *Client) logRetrySuccess(operation string, totalAttempts int) {
	if c.logger.GetSink() != nil {
		c.logger.Info("API operation succeeded after retries",
			"operation", operation,
			"attempts", totalAttempts)
	}
}