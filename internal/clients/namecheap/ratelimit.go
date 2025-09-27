package namecheap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages API rate limiting to prevent hitting Namecheap limits
type RateLimiter struct {
	limiter    *rate.Limiter
	maxRetries int
	retryDelay time.Duration
	mu         sync.RWMutex
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	// RequestsPerSecond limits the rate of API calls
	RequestsPerSecond float64
	// BurstSize allows temporary bursts above the rate limit
	BurstSize int
	// MaxRetries for rate limit exceeded errors
	MaxRetries int
	// RetryDelay base delay when rate limited
	RetryDelay time.Duration
}

// DefaultRateLimitConfig returns conservative defaults based on Namecheap API limits
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 2.0, // Conservative: 2 RPS (Namecheap allows ~20/min)
		BurstSize:         5,   // Allow small bursts
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
	}
}

// NewRateLimiter creates a new rate limiter with the given config
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		limiter:    rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize),
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
	}
}

// Wait blocks until the rate limiter allows the request
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()

	return limiter.Wait(ctx)
}

// Allow checks if a request is allowed without blocking
func (rl *RateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.limiter.Allow()
}

// UpdateLimit dynamically adjusts the rate limit
func (rl *RateLimiter) UpdateLimit(requestsPerSecond float64, burstSize int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limiter.SetLimit(rate.Limit(requestsPerSecond))
	rl.limiter.SetBurst(burstSize)
}

// GetCurrentLimit returns the current rate limit settings
func (rl *RateLimiter) GetCurrentLimit() (float64, int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return float64(rl.limiter.Limit()), rl.limiter.Burst()
}

// CircuitBreaker implements circuit breaker pattern for API calls
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	mu           sync.RWMutex
	failures     int
	lastFailTime time.Time
	state        CircuitState
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// CircuitClosed - normal operation
	CircuitClosed CircuitState = iota
	// CircuitOpen - circuit is open, requests fail fast
	CircuitOpen
	// CircuitHalfOpen - testing if service has recovered
	CircuitHalfOpen
)

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures  int
	ResetTimeout time.Duration
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  config.MaxFailures,
		resetTimeout: config.ResetTimeout,
		state:        CircuitClosed,
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	failures := cb.failures
	lastFailTime := cb.lastFailTime
	cb.mu.RUnlock()

	// Check if we should transition from Open to Half-Open
	if state == CircuitOpen && time.Since(lastFailTime) > cb.resetTimeout {
		cb.mu.Lock()
		if cb.state == CircuitOpen && time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
		}
		state = cb.state
		cb.mu.Unlock()
	}

	// Fail fast if circuit is open
	if state == CircuitOpen {
		return fmt.Errorf("circuit breaker is open (%d failures, last: %v ago)",
			failures, time.Since(lastFailTime))
	}

	// Execute the function
	err := fn()

	// Update circuit breaker state based on result
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		// Transition to Open if we've exceeded max failures
		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		} else if cb.state == CircuitHalfOpen {
			// Failed in half-open state, go back to open
			cb.state = CircuitOpen
		}

		return err
	}

	// Success - reset circuit breaker
	if cb.state == CircuitHalfOpen || cb.failures > 0 {
		cb.state = CircuitClosed
		cb.failures = 0
	}

	return nil
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() (CircuitState, int, time.Time) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state, cb.failures, cb.lastFailTime
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failures = 0
	cb.lastFailTime = time.Time{}
}