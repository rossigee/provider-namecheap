package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookServer(t *testing.T) {
	logger := logr.Discard()
	secret := "test-secret-key"

	config := Config{
		Port:         8080,
		Path:         "/webhook",
		Secret:       secret,
		Logger:       logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)

	// Register a test processor
	processed := false
	testProcessor := EventProcessorFunc(func(ctx context.Context, event *WebhookEvent) error {
		processed = true
		assert.Equal(t, "test-event-id", event.ID)
		assert.Equal(t, EventDomainRegistered, event.Type)
		return nil
	})

	server.RegisterProcessor(EventDomainRegistered, testProcessor)

	t.Run("valid webhook request", func(t *testing.T) {
		processed = false

		// Create test event
		event := WebhookEvent{
			ID:        "test-event-id",
			Type:      EventDomainRegistered,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"domain": "example.com",
			},
		}

		body, err := json.Marshal(event)
		require.NoError(t, err)

		// Generate signature
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		// Create request
		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-Namecheap-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, processed, "Event should have been processed")

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
		assert.Equal(t, "test-event-id", response["id"])
	})

	t.Run("invalid signature", func(t *testing.T) {
		processed = false

		event := WebhookEvent{
			ID:        "test-event-id",
			Type:      EventDomainRegistered,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"domain": "example.com"},
		}

		body, err := json.Marshal(event)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-Namecheap-Signature", "invalid-signature")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, processed, "Event should not have been processed")
	})

	t.Run("missing signature", func(t *testing.T) {
		processed = false

		event := WebhookEvent{
			ID:        "test-event-id",
			Type:      EventDomainRegistered,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"domain": "example.com"},
		}

		body, err := json.Marshal(event)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, processed, "Event should not have been processed")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		processed = false

		body := []byte("invalid json")

		// Generate signature for invalid JSON
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-Namecheap-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleWebhook(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, processed, "Event should not have been processed")
	})

	t.Run("unregistered event type", func(t *testing.T) {
		processed = false

		event := WebhookEvent{
			ID:        "test-event-id",
			Type:      EventDNSRecordCreated, // Not registered
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"domain": "example.com"},
		}

		body, err := json.Marshal(event)
		require.NoError(t, err)

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-Namecheap-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, processed, "Event should not have been processed")
	})

	t.Run("processor error", func(t *testing.T) {
		// Register a processor that returns an error
		errorProcessor := EventProcessorFunc(func(ctx context.Context, event *WebhookEvent) error {
			return fmt.Errorf("processing failed")
		})

		server.RegisterProcessor(EventDNSRecordCreated, errorProcessor)

		event := WebhookEvent{
			ID:        "test-event-id",
			Type:      EventDNSRecordCreated,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"domain": "example.com"},
		}

		body, err := json.Marshal(event)
		require.NoError(t, err)

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-Namecheap-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleWebhook(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("health endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.handleHealth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var health map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health["status"])
		assert.Contains(t, health, "timestamp")
		assert.Contains(t, health, "processors")
	})

	t.Run("metrics endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		server.handleMetrics(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var metrics map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &metrics)
		require.NoError(t, err)

		// Check that metrics are present
		assert.Contains(t, metrics, "requests_total")
		assert.Contains(t, metrics, "requests_errors")
		assert.Contains(t, metrics, "events_processed")
	})
}

func TestSignatureVerification(t *testing.T) {
	logger := logr.Discard()

	t.Run("no secret configured", func(t *testing.T) {
		config := Config{
			Port:   8080,
			Path:   "/webhook",
			Secret: "", // No secret
			Logger: logger,
		}

		server := NewServer(config)
		assert.True(t, server.verifySignature([]byte("test"), "any-signature"))
	})

	t.Run("valid signature", func(t *testing.T) {
		secret := "test-secret"
		body := []byte("test message")

		config := Config{
			Port:   8080,
			Path:   "/webhook",
			Secret: secret,
			Logger: logger,
		}

		server := NewServer(config)

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))

		assert.True(t, server.verifySignature(body, signature))
		assert.True(t, server.verifySignature(body, "sha256="+signature))
	})

	t.Run("invalid signature", func(t *testing.T) {
		secret := "test-secret"
		body := []byte("test message")

		config := Config{
			Port:   8080,
			Path:   "/webhook",
			Secret: secret,
			Logger: logger,
		}

		server := NewServer(config)

		assert.False(t, server.verifySignature(body, "invalid"))
		assert.False(t, server.verifySignature(body, ""))
	})
}

func TestMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test counter
	assert.Equal(t, int64(0), metrics.RequestsTotal.Value())
	metrics.RequestsTotal.Inc()
	assert.Equal(t, int64(1), metrics.RequestsTotal.Value())
	metrics.RequestsTotal.Add(5)
	assert.Equal(t, int64(6), metrics.RequestsTotal.Value())

	// Test histogram
	assert.Equal(t, int64(0), metrics.RequestDuration.Count())
	assert.Equal(t, float64(0), metrics.RequestDuration.Average())

	metrics.RequestDuration.Observe(1.0)
	metrics.RequestDuration.Observe(2.0)
	metrics.RequestDuration.Observe(3.0)

	assert.Equal(t, int64(3), metrics.RequestDuration.Count())
	assert.Equal(t, float64(2.0), metrics.RequestDuration.Average())

	// Test GetAll
	all := metrics.GetAll()
	assert.Equal(t, int64(6), all["requests_total"])
	assert.Equal(t, int64(3), all["request_count"])
	assert.Equal(t, float64(2.0), all["request_duration_avg"])
	assert.Contains(t, all, "uptime_seconds")
	assert.Contains(t, all, "last_reset")

	// Test Reset
	metrics.Reset()
	assert.Equal(t, int64(0), metrics.RequestsTotal.Value())
	assert.Equal(t, int64(0), metrics.RequestDuration.Count())
}

func TestEventProcessors(t *testing.T) {
	logger := logr.Discard()

	t.Run("domain processor", func(t *testing.T) {
		processor := NewDomainEventProcessor(logger)

		event := &WebhookEvent{
			ID:        "test-id",
			Type:      EventDomainRegistered,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"domain": "example.com",
			},
		}

		err := processor.Process(context.Background(), event)
		assert.NoError(t, err)

		// Test with missing domain
		event.Data = map[string]interface{}{}
		err = processor.Process(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing or invalid domain field")

		// Test with unsupported event type
		event.Type = "unsupported.event"
		event.Data = map[string]interface{}{"domain": "example.com"}
		err = processor.Process(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported domain event type")
	})

	t.Run("dns processor", func(t *testing.T) {
		processor := NewDNSEventProcessor(logger)

		event := &WebhookEvent{
			ID:        "test-id",
			Type:      EventDNSRecordCreated,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"domain": "example.com",
				"record": map[string]interface{}{
					"type":  "A",
					"name":  "www",
					"value": "192.168.1.1",
				},
			},
		}

		err := processor.Process(context.Background(), event)
		assert.NoError(t, err)

		// Test with missing record
		event.Data = map[string]interface{}{"domain": "example.com"}
		err = processor.Process(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing or invalid record field")
	})

	t.Run("logging processor", func(t *testing.T) {
		processor := NewLoggingEventProcessor(logger)

		event := &WebhookEvent{
			ID:        "test-id",
			Type:      EventDomainRegistered,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"domain": "example.com",
			},
		}

		err := processor.Process(context.Background(), event)
		assert.NoError(t, err)
	})
}