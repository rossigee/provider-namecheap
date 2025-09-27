package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
)

// WebhookConfig represents the configuration for webhook endpoints
type WebhookConfig struct {
	// Endpoint configuration
	URL              string        `json:"url"`
	Secret           string        `json:"secret"`
	Events           []EventType   `json:"events"`
	Active           bool          `json:"active"`

	// HTTP configuration
	Timeout          time.Duration `json:"timeout"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`

	// Security configuration
	VerifySSL        bool          `json:"verify_ssl"`
	UserAgent        string        `json:"user_agent"`

	// Metadata
	Description      string        `json:"description"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// DefaultWebhookConfig returns sensible defaults for webhook configuration
func DefaultWebhookConfig() WebhookConfig {
	return WebhookConfig{
		Active:     true,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
		VerifySSL:  true,
		UserAgent:  "crossplane-provider-namecheap/1.0",
		Events: []EventType{
			EventDomainRegistered,
			EventDomainRenewed,
			EventDomainExpired,
			EventDNSRecordCreated,
			EventDNSRecordUpdated,
			EventDNSRecordDeleted,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// WebhookManager manages webhook configurations and processors
type WebhookManager struct {
	server     *Server
	logger     logr.Logger
	processors map[EventType][]EventProcessor
}

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(server *Server, logger logr.Logger) *WebhookManager {
	return &WebhookManager{
		server:     server,
		logger:     logger.WithName("webhook-manager"),
		processors: make(map[EventType][]EventProcessor),
	}
}

// RegisterDefaultProcessors registers the default event processors
func (wm *WebhookManager) RegisterDefaultProcessors() {
	// Domain event processors
	domainProcessor := NewDomainEventProcessor(wm.logger)
	wm.server.RegisterProcessor(EventDomainRegistered, domainProcessor)
	wm.server.RegisterProcessor(EventDomainRenewed, domainProcessor)
	wm.server.RegisterProcessor(EventDomainExpired, domainProcessor)
	wm.server.RegisterProcessor(EventDomainTransferred, domainProcessor)

	// DNS event processors
	dnsProcessor := NewDNSEventProcessor(wm.logger)
	wm.server.RegisterProcessor(EventDNSRecordCreated, dnsProcessor)
	wm.server.RegisterProcessor(EventDNSRecordUpdated, dnsProcessor)
	wm.server.RegisterProcessor(EventDNSRecordDeleted, dnsProcessor)

	// SSL event processors
	sslProcessor := NewSSLEventProcessor(wm.logger)
	wm.server.RegisterProcessor(EventSSLIssued, sslProcessor)
	wm.server.RegisterProcessor(EventSSLRenewed, sslProcessor)
	wm.server.RegisterProcessor(EventSSLExpired, sslProcessor)
	wm.server.RegisterProcessor(EventSSLRevoked, sslProcessor)

	// Account event processors
	accountProcessor := NewAccountEventProcessor(wm.logger)
	wm.server.RegisterProcessor(EventAccountUpdated, accountProcessor)
	wm.server.RegisterProcessor(EventPaymentReceived, accountProcessor)
	wm.server.RegisterProcessor(EventPaymentFailed, accountProcessor)

	// Logging processor for all events (for debugging)
	loggingProcessor := NewLoggingEventProcessor(wm.logger)
	for _, eventType := range []EventType{
		EventDomainRegistered, EventDomainRenewed, EventDomainExpired, EventDomainTransferred,
		EventDNSRecordCreated, EventDNSRecordUpdated, EventDNSRecordDeleted,
		EventSSLIssued, EventSSLRenewed, EventSSLExpired, EventSSLRevoked,
		EventAccountUpdated, EventPaymentReceived, EventPaymentFailed,
	} {
		wm.AddProcessor(eventType, loggingProcessor)
	}

	wm.logger.Info("Default webhook processors registered")
}

// AddProcessor adds an additional processor for an event type
func (wm *WebhookManager) AddProcessor(eventType EventType, processor EventProcessor) {
	wm.processors[eventType] = append(wm.processors[eventType], processor)
	wm.logger.Info("Added additional processor", "event_type", eventType)
}

// RemoveProcessor removes a processor for an event type
func (wm *WebhookManager) RemoveProcessor(eventType EventType, processor EventProcessor) {
	processors := wm.processors[eventType]
	for i, p := range processors {
		if p == processor {
			wm.processors[eventType] = append(processors[:i], processors[i+1:]...)
			wm.logger.Info("Removed processor", "event_type", eventType)
			break
		}
	}
}

// GetProcessors returns all processors for an event type
func (wm *WebhookManager) GetProcessors(eventType EventType) []EventProcessor {
	return wm.processors[eventType]
}

// ValidateConfig validates webhook configuration
func ValidateConfig(config WebhookConfig) error {
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("webhook timeout must be positive")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if len(config.Events) == 0 {
		return fmt.Errorf("at least one event type must be specified")
	}

	// Validate event types
	validEvents := map[EventType]bool{
		EventDomainRegistered:  true,
		EventDomainRenewed:     true,
		EventDomainExpired:     true,
		EventDomainTransferred: true,
		EventDNSRecordCreated:  true,
		EventDNSRecordUpdated:  true,
		EventDNSRecordDeleted:  true,
		EventSSLIssued:         true,
		EventSSLRenewed:        true,
		EventSSLExpired:        true,
		EventSSLRevoked:        true,
		EventAccountUpdated:    true,
		EventPaymentReceived:   true,
		EventPaymentFailed:     true,
	}

	for _, event := range config.Events {
		if !validEvents[event] {
			return fmt.Errorf("invalid event type: %s", event)
		}
	}

	return nil
}

// WebhookSetup provides utilities for setting up webhooks
type WebhookSetup struct {
	logger logr.Logger
}

// NewWebhookSetup creates a new webhook setup utility
func NewWebhookSetup(logger logr.Logger) *WebhookSetup {
	return &WebhookSetup{
		logger: logger.WithName("webhook-setup"),
	}
}

// SetupWebhookServer creates and configures a complete webhook server
func (ws *WebhookSetup) SetupWebhookServer(config Config) (*Server, *WebhookManager, error) {
	// Create webhook server
	server := NewServer(config)

	// Create webhook manager
	manager := NewWebhookManager(server, ws.logger)

	// Register default processors
	manager.RegisterDefaultProcessors()

	ws.logger.Info("Webhook server setup complete",
		"port", config.Port,
		"path", config.Path,
		"tls_enabled", config.TLSCertFile != "" && config.TLSKeyFile != "")

	return server, manager, nil
}

// StartWebhookServer starts the webhook server with proper lifecycle management
func (ws *WebhookSetup) StartWebhookServer(ctx context.Context, server *Server, config Config) error {
	ws.logger.Info("Starting webhook server",
		"addr", fmt.Sprintf(":%d", config.Port),
		"path", config.Path)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(ctx, config.TLSCertFile, config.TLSKeyFile); err != nil {
			errChan <- fmt.Errorf("webhook server failed to start: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		ws.logger.Info("Shutting down webhook server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return server.Stop(shutdownCtx)
	}
}