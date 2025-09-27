package namecheap

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// Client represents a Namecheap API client
type Client struct {
	apiUser         string
	apiKey          string
	username        string
	clientIP        string
	baseURL         string
	httpClient      *http.Client
	sandbox         bool
	logger          logr.Logger
	rateLimiter     *RateLimiter
	circuitBreaker  *CircuitBreaker
	retryConfig     *RetryConfig
}

// Config holds the configuration for the Namecheap client
type Config struct {
	APIUser               string
	APIKey                string
	Username              string
	ClientIP              string
	BaseURL               string
	Sandbox               bool
	HTTPClient            *http.Client
	Logger                logr.Logger
	RateLimitConfig       *RateLimitConfig
	CircuitBreakerConfig  *CircuitBreakerConfig
	RetryConfig           *RetryConfig
}

// NewClient creates a new Namecheap API client
func NewClient(config Config) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://api.sandbox.namecheap.com/xml.response"
		} else {
			config.BaseURL = "https://api.namecheap.com/xml.response"
		}
	}

	// Initialize production hardening features with defaults if not provided
	rateLimitConfig := config.RateLimitConfig
	if rateLimitConfig == nil {
		defaultConfig := DefaultRateLimitConfig()
		rateLimitConfig = &defaultConfig
	}

	circuitBreakerConfig := config.CircuitBreakerConfig
	if circuitBreakerConfig == nil {
		defaultConfig := DefaultCircuitBreakerConfig()
		circuitBreakerConfig = &defaultConfig
	}

	retryConfig := config.RetryConfig
	if retryConfig == nil {
		defaultConfig := DefaultRetryConfig()
		retryConfig = &defaultConfig
	}

	return &Client{
		apiUser:         config.APIUser,
		apiKey:          config.APIKey,
		username:        config.Username,
		clientIP:        config.ClientIP,
		baseURL:         config.BaseURL,
		httpClient:      config.HTTPClient,
		sandbox:         config.Sandbox,
		logger:          config.Logger,
		rateLimiter:     NewRateLimiter(*rateLimitConfig),
		circuitBreaker:  NewCircuitBreaker(*circuitBreakerConfig),
		retryConfig:     retryConfig,
	}
}

// APIResponse represents the base structure of Namecheap API responses
type APIResponse struct {
	XMLName xml.Name `xml:"ApiResponse"`
	Status  string   `xml:"Status,attr"`
	Errors  []Error  `xml:"Errors>Error"`
}

// Error represents an API error
type Error struct {
	Number      string `xml:"Number,attr"`
	Description string `xml:",chardata"`
}

// Error implements the error interface
func (e Error) Error() string {
	return fmt.Sprintf("Namecheap API Error %s: %s", e.Number, e.Description)
}

// makeRequest performs an API request to Namecheap with production hardening
func (c *Client) makeRequest(ctx context.Context, command string, params map[string]string) (*http.Response, error) {
	var resp *http.Response

	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(err, "rate limit exceeded")
	}

	// Execute with circuit breaker and retry logic
	err := c.circuitBreaker.Execute(ctx, func() error {
		return c.WithRetry(ctx, command, func(ctx context.Context) error {
			var err error
			resp, err = c.doHTTPRequest(ctx, command, params)
			return err
		})
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// doHTTPRequest performs the actual HTTP request
func (c *Client) doHTTPRequest(ctx context.Context, command string, params map[string]string) (*http.Response, error) {
	values := url.Values{}
	values.Set("ApiUser", c.apiUser)
	values.Set("ApiKey", c.apiKey)
	values.Set("UserName", c.username)
	values.Set("ClientIp", c.clientIP)
	values.Set("Command", command)

	// Add additional parameters
	for key, value := range params {
		values.Set(key, value)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.URL.RawQuery = values.Encode()
	req.Header.Set("User-Agent", "crossplane-provider-namecheap/1.0")

	if c.logger.Enabled() {
		c.logger.V(1).Info("Making API request",
			"command", command,
			"url", req.URL.String())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}

	// Check for HTTP-level errors that should trigger retries
	if resp.StatusCode >= 500 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Server error: %s", resp.Status),
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    "Rate limit exceeded",
		}
	}

	return resp, nil
}

// parseResponse parses the API response and checks for errors
func parseResponse(resp *http.Response, result interface{}) error {
	defer func() {
		_ = resp.Body.Close() // Ignore close errors
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// First parse the base response to check for API errors
	var baseResp APIResponse
	if err := xml.Unmarshal(body, &baseResp); err != nil {
		return errors.Wrap(err, "failed to parse API response")
	}

	if baseResp.Status != "OK" {
		if len(baseResp.Errors) > 0 {
			return baseResp.Errors[0]
		}
		return errors.New("API request failed with unknown error")
	}

	// Parse the full response into the result struct
	if err := xml.Unmarshal(body, result); err != nil {
		return errors.Wrap(err, "failed to parse response into result struct")
	}

	return nil
}