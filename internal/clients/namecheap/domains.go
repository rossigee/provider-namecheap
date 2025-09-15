package namecheap

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Domain represents a domain in Namecheap
type Domain struct {
	ID             int       `xml:"ID,attr"`
	Name           string    `xml:"Name,attr"`
	User           string    `xml:"User,attr"`
	Created        time.Time `xml:"Created,attr"`
	Expires        time.Time `xml:"Expires,attr"`
	IsExpired      bool      `xml:"IsExpired,attr"`
	IsLocked       bool      `xml:"IsLocked,attr"`
	AutoRenew      bool      `xml:"AutoRenew,attr"`
	WhoisGuard     string    `xml:"WhoisGuard,attr"`
	IsPremium      bool      `xml:"IsPremium,attr"`
	IsOurDNS       bool      `xml:"IsOurDNS,attr"`
}

// DomainListResponse represents the response from domains.getList
type DomainListResponse struct {
	APIResponse
	CommandResponse struct {
		DomainGetListResult struct {
			Domains []Domain `xml:"Domain"`
		} `xml:"DomainGetListResult"`
	} `xml:"CommandResponse"`
}

// DomainInfoResponse represents the response from domains.getInfo
type DomainInfoResponse struct {
	APIResponse
	CommandResponse struct {
		DomainGetInfoResult struct {
			Domain     Domain `xml:"DomainDetails"`
			DnsDetails struct {
				ProviderType  string   `xml:"ProviderType,attr"`
				IsUsingOurDNS bool     `xml:"IsUsingOurDNS,attr"`
				Nameservers   []string `xml:"Nameserver"`
			} `xml:"DnsDetails"`
		} `xml:"DomainGetInfoResult"`
	} `xml:"CommandResponse"`
}

// DomainCreateResponse represents the response from domains.create
type DomainCreateResponse struct {
	APIResponse
	CommandResponse struct {
		DomainCreateResult struct {
			Domain                 string  `xml:"Domain,attr"`
			Registered             bool    `xml:"Registered,attr"`
			ChargedAmount          float64 `xml:"ChargedAmount,attr"`
			DomainID               int     `xml:"DomainID,attr"`
			OrderID                int     `xml:"OrderID,attr"`
			TransactionID          int     `xml:"TransactionID,attr"`
			WhoisGuardEnable       bool    `xml:"WhoisguardEnable,attr"`
			NonRealTimeDomain      bool    `xml:"NonRealTimeDomain,attr"`
		} `xml:"DomainCreateResult"`
	} `xml:"CommandResponse"`
}

// DNSSetCustomResponse represents the response from domains.dns.setCustom
type DNSSetCustomResponse struct {
	APIResponse
	CommandResponse struct {
		DomainDNSSetCustomResult struct {
			Domain  string `xml:"Domain,attr"`
			Updated bool   `xml:"Updated,attr"`
		} `xml:"DomainDNSSetCustomResult"`
	} `xml:"CommandResponse"`
}

// GetDomains retrieves a list of domains for the account
func (c *Client) GetDomains(ctx context.Context) ([]Domain, error) {
	resp, err := c.makeRequest(ctx, "namecheap.domains.getList", map[string]string{
		"PageSize": "100",
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.getList request")
	}

	var result DomainListResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.getList response")
	}

	return result.CommandResponse.DomainGetListResult.Domains, nil
}

// GetDomain retrieves detailed information about a specific domain
func (c *Client) GetDomain(ctx context.Context, domainName string) (*Domain, error) {
	resp, err := c.makeRequest(ctx, "namecheap.domains.getInfo", map[string]string{
		"DomainName": domainName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.getInfo request")
	}

	var result DomainInfoResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.getInfo response")
	}

	domain := result.CommandResponse.DomainGetInfoResult.Domain
	return &domain, nil
}

// CreateDomain registers a new domain
func (c *Client) CreateDomain(ctx context.Context, domainName string, years int) (*Domain, error) {
	params := map[string]string{
		"DomainName": domainName,
		"Years":      strconv.Itoa(years),
	}

	resp, err := c.makeRequest(ctx, "namecheap.domains.create", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.create request")
	}

	var result DomainCreateResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.create response")
	}

	if !result.CommandResponse.DomainCreateResult.Registered {
		return nil, errors.New("domain registration failed")
	}

	// After registration, get the domain details
	return c.GetDomain(ctx, domainName)
}

// SetNameservers sets custom nameservers for a domain
func (c *Client) SetNameservers(ctx context.Context, domainName string, nameservers []string) error {
	if len(nameservers) == 0 {
		return errors.New("at least one nameserver must be provided")
	}

	params := map[string]string{
		"SLD": strings.Split(domainName, ".")[0],
		"TLD": strings.Join(strings.Split(domainName, ".")[1:], "."),
		"Nameservers": strings.Join(nameservers, ","),
	}

	resp, err := c.makeRequest(ctx, "namecheap.domains.dns.setCustom", params)
	if err != nil {
		return errors.Wrap(err, "failed to make domains.dns.setCustom request")
	}

	var result DNSSetCustomResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse domains.dns.setCustom response")
	}

	if !result.CommandResponse.DomainDNSSetCustomResult.Updated {
		return errors.New("failed to update nameservers")
	}

	return nil
}

// DomainRenewResponse represents the response from domains.renew
type DomainRenewResponse struct {
	APIResponse
	CommandResponse struct {
		DomainRenewResult struct {
			DomainName    string  `xml:"DomainName,attr"`
			DomainID      int     `xml:"DomainID,attr"`
			Renew         bool    `xml:"Renew,attr"`
			ChargedAmount float64 `xml:"ChargedAmount,attr"`
			TransactionID int     `xml:"TransactionID,attr"`
			OrderID       int     `xml:"OrderID,attr"`
		} `xml:"DomainRenewResult"`
	} `xml:"CommandResponse"`
}

// DomainCheckResponse represents the response from domains.check
type DomainCheckResponse struct {
	APIResponse
	CommandResponse struct {
		DomainCheckResult struct {
			Domains []struct {
				Domain      string `xml:"Domain,attr"`
				Available   bool   `xml:"Available,attr"`
				ErrorCode   string `xml:"ErrorCode,attr"`
				Description string `xml:"Description,attr"`
				IsPremium   bool   `xml:"IsPremium,attr"`
				PremiumRegistrationPrice float64 `xml:"PremiumRegistrationPrice,attr"`
				PremiumRenewalPrice      float64 `xml:"PremiumRenewalPrice,attr"`
				PremiumRestorePrice      float64 `xml:"PremiumRestorePrice,attr"`
				PremiumTransferPrice     float64 `xml:"PremiumTransferPrice,attr"`
				IcannFee                 float64 `xml:"IcannFee,attr"`
				EapFee                   float64 `xml:"EapFee,attr"`
			} `xml:"DomainCheckResult"`
		} `xml:"DomainCheckResult"`
	} `xml:"CommandResponse"`
}

// DomainCheckResult represents a single domain availability check result
type DomainCheckResult struct {
	Domain      string
	Available   bool
	ErrorCode   string
	Description string
	IsPremium   bool
	PremiumRegistrationPrice float64
	PremiumRenewalPrice      float64
	PremiumRestorePrice      float64
	PremiumTransferPrice     float64
	IcannFee                 float64
	EapFee                   float64
}

// RenewDomain renews a domain for specified number of years
func (c *Client) RenewDomain(ctx context.Context, domainName string, years int) (*Domain, error) {
	params := map[string]string{
		"DomainName": domainName,
		"Years":      strconv.Itoa(years),
	}

	resp, err := c.makeRequest(ctx, "namecheap.domains.renew", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.renew request")
	}

	var result DomainRenewResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.renew response")
	}

	if !result.CommandResponse.DomainRenewResult.Renew {
		return nil, errors.New("domain renewal failed")
	}

	// After renewal, get the updated domain details
	return c.GetDomain(ctx, domainName)
}

// CheckDomainAvailability checks if domains are available for registration
func (c *Client) CheckDomainAvailability(ctx context.Context, domainNames []string) ([]DomainCheckResult, error) {
	if len(domainNames) == 0 {
		return nil, errors.New("at least one domain name must be provided")
	}

	params := map[string]string{
		"DomainList": strings.Join(domainNames, ","),
	}

	resp, err := c.makeRequest(ctx, "namecheap.domains.check", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.check request")
	}

	var result DomainCheckResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.check response")
	}

	// Convert API response to our result type
	checkResults := make([]DomainCheckResult, len(result.CommandResponse.DomainCheckResult.Domains))
	for i, domain := range result.CommandResponse.DomainCheckResult.Domains {
		checkResults[i] = DomainCheckResult{
			Domain:                   domain.Domain,
			Available:                domain.Available,
			ErrorCode:                domain.ErrorCode,
			Description:              domain.Description,
			IsPremium:                domain.IsPremium,
			PremiumRegistrationPrice: domain.PremiumRegistrationPrice,
			PremiumRenewalPrice:      domain.PremiumRenewalPrice,
			PremiumRestorePrice:      domain.PremiumRestorePrice,
			PremiumTransferPrice:     domain.PremiumTransferPrice,
			IcannFee:                 domain.IcannFee,
			EapFee:                   domain.EapFee,
		}
	}

	return checkResults, nil
}

// DomainExists checks if a domain exists in the account
func (c *Client) DomainExists(ctx context.Context, domainName string) (bool, error) {
	_, err := c.GetDomain(ctx, domainName)
	if err != nil {
		// Check if it's a "domain not found" error
		if strings.Contains(err.Error(), "Domain not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}