package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
)

// DomainEventProcessor handles domain-related webhook events
type DomainEventProcessor struct {
	logger logr.Logger
}

// NewDomainEventProcessor creates a new domain event processor
func NewDomainEventProcessor(logger logr.Logger) *DomainEventProcessor {
	return &DomainEventProcessor{
		logger: logger.WithName("domain-processor"),
	}
}

// Process handles domain events (registered, renewed, expired, transferred)
func (p *DomainEventProcessor) Process(ctx context.Context, event *WebhookEvent) error {
	p.logger.Info("Processing domain event",
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp)

	// Extract domain information from event data
	domainName, ok := event.Data["domain"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid domain field in event data")
	}

	switch event.Type {
	case EventDomainRegistered:
		return p.handleDomainRegistered(ctx, domainName, event.Data)
	case EventDomainRenewed:
		return p.handleDomainRenewed(ctx, domainName, event.Data)
	case EventDomainExpired:
		return p.handleDomainExpired(ctx, domainName, event.Data)
	case EventDomainTransferred:
		return p.handleDomainTransferred(ctx, domainName, event.Data)
	default:
		return fmt.Errorf("unsupported domain event type: %s", event.Type)
	}
}

func (p *DomainEventProcessor) handleDomainRegistered(ctx context.Context, domain string, data map[string]interface{}) error {
	p.logger.Info("Domain registered successfully", "domain", domain)
	// Here you could update the domain resource status in Kubernetes
	// or trigger additional provisioning workflows
	return nil
}

func (p *DomainEventProcessor) handleDomainRenewed(ctx context.Context, domain string, data map[string]interface{}) error {
	p.logger.Info("Domain renewed", "domain", domain)
	if expiryDate, ok := data["expiry_date"].(string); ok {
		p.logger.Info("Domain renewal details", "domain", domain, "new_expiry", expiryDate)
	}
	return nil
}

func (p *DomainEventProcessor) handleDomainExpired(ctx context.Context, domain string, data map[string]interface{}) error {
	p.logger.Error(nil, "Domain expired", "domain", domain)
	// Could trigger alerts or automatic renewal workflows
	return nil
}

func (p *DomainEventProcessor) handleDomainTransferred(ctx context.Context, domain string, data map[string]interface{}) error {
	p.logger.Info("Domain transferred", "domain", domain)
	return nil
}

// DNSEventProcessor handles DNS record webhook events
type DNSEventProcessor struct {
	logger logr.Logger
}

// NewDNSEventProcessor creates a new DNS event processor
func NewDNSEventProcessor(logger logr.Logger) *DNSEventProcessor {
	return &DNSEventProcessor{
		logger: logger.WithName("dns-processor"),
	}
}

// Process handles DNS events (record created, updated, deleted)
func (p *DNSEventProcessor) Process(ctx context.Context, event *WebhookEvent) error {
	p.logger.Info("Processing DNS event",
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp)

	// Extract DNS record information
	recordData, ok := event.Data["record"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid record field in event data")
	}

	recordType, _ := recordData["type"].(string)
	recordName, _ := recordData["name"].(string)
	recordValue, _ := recordData["value"].(string)
	domain, _ := event.Data["domain"].(string)

	switch event.Type {
	case EventDNSRecordCreated:
		return p.handleRecordCreated(ctx, domain, recordType, recordName, recordValue, recordData)
	case EventDNSRecordUpdated:
		return p.handleRecordUpdated(ctx, domain, recordType, recordName, recordValue, recordData)
	case EventDNSRecordDeleted:
		return p.handleRecordDeleted(ctx, domain, recordType, recordName, recordData)
	default:
		return fmt.Errorf("unsupported DNS event type: %s", event.Type)
	}
}

func (p *DNSEventProcessor) handleRecordCreated(ctx context.Context, domain, recordType, name, value string, data map[string]interface{}) error {
	p.logger.Info("DNS record created",
		"domain", domain,
		"type", recordType,
		"name", name,
		"value", value)
	return nil
}

func (p *DNSEventProcessor) handleRecordUpdated(ctx context.Context, domain, recordType, name, value string, data map[string]interface{}) error {
	p.logger.Info("DNS record updated",
		"domain", domain,
		"type", recordType,
		"name", name,
		"value", value)
	return nil
}

func (p *DNSEventProcessor) handleRecordDeleted(ctx context.Context, domain, recordType, name string, data map[string]interface{}) error {
	p.logger.Info("DNS record deleted",
		"domain", domain,
		"type", recordType,
		"name", name)
	return nil
}

// SSLEventProcessor handles SSL certificate webhook events
type SSLEventProcessor struct {
	logger logr.Logger
}

// NewSSLEventProcessor creates a new SSL event processor
func NewSSLEventProcessor(logger logr.Logger) *SSLEventProcessor {
	return &SSLEventProcessor{
		logger: logger.WithName("ssl-processor"),
	}
}

// Process handles SSL events (issued, renewed, expired, revoked)
func (p *SSLEventProcessor) Process(ctx context.Context, event *WebhookEvent) error {
	p.logger.Info("Processing SSL event",
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp)

	// Extract SSL certificate information
	certID, _ := event.Data["certificate_id"].(string)
	domain, _ := event.Data["domain"].(string)

	switch event.Type {
	case EventSSLIssued:
		return p.handleSSLIssued(ctx, certID, domain, event.Data)
	case EventSSLRenewed:
		return p.handleSSLRenewed(ctx, certID, domain, event.Data)
	case EventSSLExpired:
		return p.handleSSLExpired(ctx, certID, domain, event.Data)
	case EventSSLRevoked:
		return p.handleSSLRevoked(ctx, certID, domain, event.Data)
	default:
		return fmt.Errorf("unsupported SSL event type: %s", event.Type)
	}
}

func (p *SSLEventProcessor) handleSSLIssued(ctx context.Context, certID, domain string, data map[string]interface{}) error {
	p.logger.Info("SSL certificate issued", "cert_id", certID, "domain", domain)
	return nil
}

func (p *SSLEventProcessor) handleSSLRenewed(ctx context.Context, certID, domain string, data map[string]interface{}) error {
	p.logger.Info("SSL certificate renewed", "cert_id", certID, "domain", domain)
	return nil
}

func (p *SSLEventProcessor) handleSSLExpired(ctx context.Context, certID, domain string, data map[string]interface{}) error {
	p.logger.Error(nil, "SSL certificate expired", "cert_id", certID, "domain", domain)
	return nil
}

func (p *SSLEventProcessor) handleSSLRevoked(ctx context.Context, certID, domain string, data map[string]interface{}) error {
	p.logger.Error(nil, "SSL certificate revoked", "cert_id", certID, "domain", domain)
	return nil
}

// AccountEventProcessor handles account and payment webhook events
type AccountEventProcessor struct {
	logger logr.Logger
}

// NewAccountEventProcessor creates a new account event processor
func NewAccountEventProcessor(logger logr.Logger) *AccountEventProcessor {
	return &AccountEventProcessor{
		logger: logger.WithName("account-processor"),
	}
}

// Process handles account events (updated, payment received, payment failed)
func (p *AccountEventProcessor) Process(ctx context.Context, event *WebhookEvent) error {
	p.logger.Info("Processing account event",
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp)

	switch event.Type {
	case EventAccountUpdated:
		return p.handleAccountUpdated(ctx, event.Data)
	case EventPaymentReceived:
		return p.handlePaymentReceived(ctx, event.Data)
	case EventPaymentFailed:
		return p.handlePaymentFailed(ctx, event.Data)
	default:
		return fmt.Errorf("unsupported account event type: %s", event.Type)
	}
}

func (p *AccountEventProcessor) handleAccountUpdated(ctx context.Context, data map[string]interface{}) error {
	p.logger.Info("Account updated", "data", data)
	return nil
}

func (p *AccountEventProcessor) handlePaymentReceived(ctx context.Context, data map[string]interface{}) error {
	amount, _ := data["amount"].(float64)
	currency, _ := data["currency"].(string)
	p.logger.Info("Payment received", "amount", amount, "currency", currency)
	return nil
}

func (p *AccountEventProcessor) handlePaymentFailed(ctx context.Context, data map[string]interface{}) error {
	amount, _ := data["amount"].(float64)
	currency, _ := data["currency"].(string)
	reason, _ := data["reason"].(string)

	p.logger.Error(nil, "Payment failed",
		"amount", amount,
		"currency", currency,
		"reason", reason)

	// Could trigger alerts or retry mechanisms
	return nil
}

// LoggingEventProcessor is a generic processor that logs all events
type LoggingEventProcessor struct {
	logger logr.Logger
}

// NewLoggingEventProcessor creates a new logging event processor
func NewLoggingEventProcessor(logger logr.Logger) *LoggingEventProcessor {
	return &LoggingEventProcessor{
		logger: logger.WithName("logging-processor"),
	}
}

// Process logs the webhook event for debugging and audit purposes
func (p *LoggingEventProcessor) Process(ctx context.Context, event *WebhookEvent) error {
	eventJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event for logging: %w", err)
	}

	p.logger.Info("Webhook event received",
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp,
		"event_data", string(eventJSON))

	return nil
}