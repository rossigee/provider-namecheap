package namecheap

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// WhoisGuard represents a WhoisGuard privacy protection service
type WhoisGuard struct {
	ID           int    `xml:"ID,attr"`
	DomainName   string `xml:"DomainName,attr"`
	Created      string `xml:"Created,attr"`
	Status       string `xml:"Status,attr"`
	EmailDetails struct {
		ForwardedTo     string `xml:"ForwardedTo,attr"`
		LastAutoEmailDate string `xml:"LastAutoEmailDate,attr"`
		AutoEmailCount  int    `xml:"AutoEmailCount,attr"`
	} `xml:"EmailDetails"`
}

// WhoisGuardListResponse represents the response from whoisguard.getList
type WhoisGuardListResponse struct {
	APIResponse
	CommandResponse struct {
		WhoisGuardGetListResult struct {
			WhoisGuards []WhoisGuard `xml:"Whoisguard"`
		} `xml:"WhoisguardGetListResult"`
	} `xml:"CommandResponse"`
}

// WhoisGuardEnableResponse represents the response from whoisguard.enable
type WhoisGuardEnableResponse struct {
	APIResponse
	CommandResponse struct {
		WhoisGuardEnableResult struct {
			Domain    string `xml:"Domain,attr"`
			IsSuccess bool   `xml:"IsSuccess,attr"`
		} `xml:"WhoisguardEnableResult"`
	} `xml:"CommandResponse"`
}

// WhoisGuardDisableResponse represents the response from whoisguard.disable
type WhoisGuardDisableResponse struct {
	APIResponse
	CommandResponse struct {
		WhoisGuardDisableResult struct {
			Domain    string `xml:"Domain,attr"`
			IsSuccess bool   `xml:"IsSuccess,attr"`
		} `xml:"WhoisguardDisableResult"`
	} `xml:"CommandResponse"`
}

// WhoisGuardRenewResponse represents the response from whoisguard.renew
type WhoisGuardRenewResponse struct {
	APIResponse
	CommandResponse struct {
		WhoisGuardRenewResult struct {
			WhoisguardID  int     `xml:"WhoisguardID,attr"`
			Renew         bool    `xml:"Renew,attr"`
			ChargedAmount float64 `xml:"ChargedAmount,attr"`
			TransactionID int     `xml:"TransactionID,attr"`
			OrderID       int     `xml:"OrderID,attr"`
		} `xml:"WhoisguardRenewResult"`
	} `xml:"CommandResponse"`
}

// GetWhoisGuards retrieves all WhoisGuard services for the account
func (c *Client) GetWhoisGuards(ctx context.Context) ([]WhoisGuard, error) {
	resp, err := c.makeRequest(ctx, "namecheap.whoisguard.getList", map[string]string{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make whoisguard.getList request")
	}

	var result WhoisGuardListResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse whoisguard.getList response")
	}

	return result.CommandResponse.WhoisGuardGetListResult.WhoisGuards, nil
}

// EnableWhoisGuard enables WhoisGuard privacy protection for a domain
func (c *Client) EnableWhoisGuard(ctx context.Context, whoisGuardID int, domainName, forwardedToEmail string) error {
	params := map[string]string{
		"WhoisguardID": strconv.Itoa(whoisGuardID),
		"DomainName":   domainName,
	}

	if forwardedToEmail != "" {
		params["ForwardedToEmail"] = forwardedToEmail
	}

	resp, err := c.makeRequest(ctx, "namecheap.whoisguard.enable", params)
	if err != nil {
		return errors.Wrap(err, "failed to make whoisguard.enable request")
	}

	var result WhoisGuardEnableResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse whoisguard.enable response")
	}

	if !result.CommandResponse.WhoisGuardEnableResult.IsSuccess {
		return errors.New("failed to enable WhoisGuard")
	}

	return nil
}

// DisableWhoisGuard disables WhoisGuard privacy protection for a domain
func (c *Client) DisableWhoisGuard(ctx context.Context, whoisGuardID int, domainName string) error {
	params := map[string]string{
		"WhoisguardID": strconv.Itoa(whoisGuardID),
		"DomainName":   domainName,
	}

	resp, err := c.makeRequest(ctx, "namecheap.whoisguard.disable", params)
	if err != nil {
		return errors.Wrap(err, "failed to make whoisguard.disable request")
	}

	var result WhoisGuardDisableResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse whoisguard.disable response")
	}

	if !result.CommandResponse.WhoisGuardDisableResult.IsSuccess {
		return errors.New("failed to disable WhoisGuard")
	}

	return nil
}

// RenewWhoisGuard renews WhoisGuard privacy protection service
func (c *Client) RenewWhoisGuard(ctx context.Context, whoisGuardID int, years int) error {
	params := map[string]string{
		"WhoisguardID": strconv.Itoa(whoisGuardID),
		"Years":        strconv.Itoa(years),
	}

	resp, err := c.makeRequest(ctx, "namecheap.whoisguard.renew", params)
	if err != nil {
		return errors.Wrap(err, "failed to make whoisguard.renew request")
	}

	var result WhoisGuardRenewResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse whoisguard.renew response")
	}

	if !result.CommandResponse.WhoisGuardRenewResult.Renew {
		return errors.New("WhoisGuard renewal failed")
	}

	return nil
}

// GetWhoisGuardForDomain retrieves WhoisGuard information for a specific domain
func (c *Client) GetWhoisGuardForDomain(ctx context.Context, domainName string) (*WhoisGuard, error) {
	whoisGuards, err := c.GetWhoisGuards(ctx)
	if err != nil {
		return nil, err
	}

	for _, wg := range whoisGuards {
		if strings.EqualFold(wg.DomainName, domainName) {
			return &wg, nil
		}
	}

	return nil, errors.New("WhoisGuard not found for domain")
}

// IsWhoisGuardEnabled checks if WhoisGuard is enabled for a domain
func (c *Client) IsWhoisGuardEnabled(ctx context.Context, domainName string) (bool, error) {
	whoisGuard, err := c.GetWhoisGuardForDomain(ctx, domainName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return whoisGuard.Status == "ENABLED", nil
}