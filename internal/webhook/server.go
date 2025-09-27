package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Server represents a webhook server for processing Namecheap events
type Server struct {
	router     *mux.Router
	server     *http.Server
	logger     logr.Logger
	secret     string
	processors map[EventType]EventProcessor
	metrics    *Metrics
}

// Config holds webhook server configuration
type Config struct {
	Port          int
	Path          string
	Secret        string
	Logger        logr.Logger
	TLSCertFile   string
	TLSKeyFile    string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

// DefaultConfig returns sensible defaults for webhook server
func DefaultConfig() Config {
	return Config{
		Port:         8443,
		Path:         "/webhook",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// EventType represents different types of Namecheap webhook events
type EventType string

const (
	// Domain events
	EventDomainRegistered   EventType = "domain.registered"
	EventDomainRenewed      EventType = "domain.renewed"
	EventDomainExpired      EventType = "domain.expired"
	EventDomainTransferred  EventType = "domain.transferred"

	// DNS events
	EventDNSRecordCreated   EventType = "dns.record.created"
	EventDNSRecordUpdated   EventType = "dns.record.updated"
	EventDNSRecordDeleted   EventType = "dns.record.deleted"

	// SSL events
	EventSSLIssued          EventType = "ssl.issued"
	EventSSLRenewed         EventType = "ssl.renewed"
	EventSSLExpired         EventType = "ssl.expired"
	EventSSLRevoked         EventType = "ssl.revoked"

	// Account events
	EventAccountUpdated     EventType = "account.updated"
	EventPaymentReceived    EventType = "payment.received"
	EventPaymentFailed      EventType = "payment.failed"
)

// WebhookEvent represents a Namecheap webhook event
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"-"` // Extracted from headers
}

// EventProcessor defines how to process different types of events
type EventProcessor interface {
	Process(ctx context.Context, event *WebhookEvent) error
}

// EventProcessorFunc allows functions to implement EventProcessor
type EventProcessorFunc func(ctx context.Context, event *WebhookEvent) error

func (f EventProcessorFunc) Process(ctx context.Context, event *WebhookEvent) error {
	return f(ctx, event)
}

// NewServer creates a new webhook server
func NewServer(config Config) *Server {
	if config.Logger.GetSink() == nil {
		config.Logger = logr.Discard()
	}

	router := mux.NewRouter()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	s := &Server{
		router:     router,
		server:     server,
		logger:     config.Logger,
		secret:     config.Secret,
		processors: make(map[EventType]EventProcessor),
		metrics:    NewMetrics(),
	}

	// Setup routes
	s.router.HandleFunc(config.Path, s.handleWebhook).Methods("POST")
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/metrics", s.handleMetrics).Methods("GET")

	return s
}

// RegisterProcessor registers an event processor for a specific event type
func (s *Server) RegisterProcessor(eventType EventType, processor EventProcessor) {
	s.processors[eventType] = processor
	s.logger.Info("Registered webhook event processor", "eventType", eventType)
}

// Start starts the webhook server
func (s *Server) Start(ctx context.Context, tlsCertFile, tlsKeyFile string) error {
	s.logger.Info("Starting webhook server", "addr", s.server.Addr)

	var err error
	if tlsCertFile != "" && tlsKeyFile != "" {
		s.logger.Info("Starting webhook server with TLS", "cert", tlsCertFile, "key", tlsKeyFile)
		err = s.server.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
	} else {
		s.logger.Info("Starting webhook server without TLS")
		err = s.server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "webhook server failed")
	}

	return nil
}

// Stop gracefully stops the webhook server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping webhook server")
	return s.server.Shutdown(ctx)
}

// handleWebhook processes incoming webhook events
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	s.metrics.RequestsTotal.Inc()
	start := time.Now()

	defer func() {
		s.metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error(err, "Failed to read webhook request body")
		s.metrics.RequestsErrors.Inc()
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify signature
	signature := r.Header.Get("X-Namecheap-Signature")
	if !s.verifySignature(body, signature) {
		s.logger.Error(nil, "Invalid webhook signature")
		s.metrics.RequestsErrors.Inc()
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse webhook event
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		s.logger.Error(err, "Failed to parse webhook event")
		s.metrics.RequestsErrors.Inc()
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	event.Signature = signature

	s.logger.Info("Received webhook event",
		"id", event.ID,
		"type", event.Type,
		"timestamp", event.Timestamp)

	// Process the event
	processor, exists := s.processors[event.Type]
	if !exists {
		s.logger.Info("No processor registered for event type", "type", event.Type)
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := processor.Process(ctx, &event); err != nil {
		s.logger.Error(err, "Failed to process webhook event",
			"id", event.ID,
			"type", event.Type)
		s.metrics.ProcessingErrors.Inc()
		http.Error(w, "Event processing failed", http.StatusInternalServerError)
		return
	}

	s.metrics.EventsProcessed.Inc()
	s.logger.Info("Successfully processed webhook event",
		"id", event.ID,
		"type", event.Type)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"status":"ok","id":"%s"}`, event.ID); err != nil {
		s.logger.Error(err, "Failed to write response")
	}
}

// verifySignature verifies the webhook signature
func (s *Server) verifySignature(body []byte, signature string) bool {
	if s.secret == "" {
		s.logger.Info("No webhook secret configured, skipping signature verification")
		return true
	}

	if signature == "" {
		return false
	}

	// Remove the "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"processors": func() []string {
			types := make([]string, 0, len(s.processors))
			for t := range s.processors {
				types = append(types, string(t))
			}
			return types
		}(),
	}); err != nil {
		s.logger.Error(err, "Failed to encode health response")
	}
}

// handleMetrics returns Prometheus metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// This would integrate with Prometheus metrics handler
	// For now, return basic metrics in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s.metrics.GetAll()); err != nil {
		s.logger.Error(err, "Failed to encode metrics response")
	}
}