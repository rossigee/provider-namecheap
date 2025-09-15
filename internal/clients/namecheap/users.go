package namecheap

import (
	"context"

	"github.com/pkg/errors"
)

// UserBalance represents account balance information
type UserBalance struct {
	Currency          string  `xml:"Currency,attr"`
	AvailableBalance  float64 `xml:"AvailableBalance,attr"`
	AccountBalance    float64 `xml:"AccountBalance,attr"`
	EarnedAmount      float64 `xml:"EarnedAmount,attr"`
	WithdrawableAmount float64 `xml:"WithdrawableAmount,attr"`
	FundsRequiredForAutoRenew float64 `xml:"FundsRequiredForAutoRenew,attr"`
}

// UserBalanceResponse represents the response from users.getBalances
type UserBalanceResponse struct {
	APIResponse
	CommandResponse struct {
		UserGetBalancesResult UserBalance `xml:"UserGetBalancesResult"`
	} `xml:"CommandResponse"`
}

// TLD represents a top-level domain with pricing information
type TLD struct {
	Name                string  `xml:"Name,attr"`
	NonRealTime         bool    `xml:"NonRealTime,attr"`
	MinRegisterYears    int     `xml:"MinRegisterYears,attr"`
	MaxRegisterYears    int     `xml:"MaxRegisterYears,attr"`
	MinRenewYears       int     `xml:"MinRenewYears,attr"`
	MaxRenewYears       int     `xml:"MaxRenewYears,attr"`
	MinTransferYears    int     `xml:"MinTransferYears,attr"`
	MaxTransferYears    int     `xml:"MaxTransferYears,attr"`
	IsApiRegisterable   bool    `xml:"IsApiRegisterable,attr"`
	IsApiRenewable      bool    `xml:"IsApiRenewable,attr"`
	IsApiTransferable   bool    `xml:"IsApiTransferable,attr"`
	IsEppRequired       bool    `xml:"IsEppRequired,attr"`
	IsDisableModContact bool    `xml:"IsDisableModContact,attr"`
	IsDisableWGAllot    bool    `xml:"IsDisableWGAllot,attr"`
	IsIncludeInExtendedSearchOnly bool `xml:"IsIncludeInExtendedSearchOnly,attr"`
	SequenceNumber      int     `xml:"SequenceNumber,attr"`
	Type                string  `xml:"Type,attr"`
	SubType             string  `xml:"SubType,attr"`
	IsSupportsIDN       bool    `xml:"IsSupportsIDN,attr"`
	Category            string  `xml:"Category,attr"`
	SupportsRegistrarLock bool  `xml:"SupportsRegistrarLock,attr"`
	AddGracePeriodFee   float64 `xml:"AddGracePeriodFee,attr"`
	WhoisVerification   bool    `xml:"WhoisVerification,attr"`
	ProviderApiDelete   bool    `xml:"ProviderApiDelete,attr"`
	TldState            string  `xml:"TldState,attr"`
	SearchGroup         string  `xml:"SearchGroup,attr"`
	Registry            string  `xml:"Registry,attr"`
}

// TLDListResponse represents the response from domains.getTldList
type TLDListResponse struct {
	APIResponse
	CommandResponse struct {
		DomainsTldListResult struct {
			TLDs []TLD `xml:"Tld"`
		} `xml:"DomainsGetTldListResult"`
	} `xml:"CommandResponse"`
}

// PricingType represents pricing information for a TLD
type PricingType struct {
	Name              string  `xml:"Name,attr"`
	Price             float64 `xml:"Price,attr"`
	RegularPrice      float64 `xml:"RegularPrice,attr"`
	YourPrice         float64 `xml:"YourPrice,attr"`
	YourPriceRange    string  `xml:"YourPriceRange,attr"`
	PromoPrice        float64 `xml:"PromoPrice,attr"`
	Currency          string  `xml:"Currency,attr"`
	Duration          int     `xml:"Duration,attr"`
	DurationType      string  `xml:"DurationType,attr"`
	PricingType       string  `xml:"PricingType,attr"`
	AdditionalCost    float64 `xml:"AdditionalCost,attr"`
}

// UserPricingResponse represents the response from users.getPricing
type UserPricingResponse struct {
	APIResponse
	CommandResponse struct {
		UserGetPricingResult struct {
			ProductType     string        `xml:"ProductType,attr"`
			ProductCategory string        `xml:"ProductCategory,attr"`
			Product         string        `xml:"Product,attr"`
			PricingTypes    []PricingType `xml:"ProductType>PricingType"`
		} `xml:"UserGetPricingResult"`
	} `xml:"CommandResponse"`
}

// GetUserBalances retrieves account balance information
func (c *Client) GetUserBalances(ctx context.Context) (*UserBalance, error) {
	resp, err := c.makeRequest(ctx, "namecheap.users.getBalances", map[string]string{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make users.getBalances request")
	}

	var result UserBalanceResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse users.getBalances response")
	}

	return &result.CommandResponse.UserGetBalancesResult, nil
}

// GetTLDList retrieves list of TLDs with their properties and capabilities
func (c *Client) GetTLDList(ctx context.Context) ([]TLD, error) {
	resp, err := c.makeRequest(ctx, "namecheap.domains.getTldList", map[string]string{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.getTldList request")
	}

	var result TLDListResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.getTldList response")
	}

	return result.CommandResponse.DomainsTldListResult.TLDs, nil
}

// GetPricing retrieves pricing information for domain registration, renewal, transfer, etc.
func (c *Client) GetPricing(ctx context.Context, productType, productCategory, action string) ([]PricingType, error) {
	params := map[string]string{
		"ProductType": productType,
		"Action":      action,
	}

	if productCategory != "" {
		params["ProductCategory"] = productCategory
	}

	resp, err := c.makeRequest(ctx, "namecheap.users.getPricing", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make users.getPricing request")
	}

	var result UserPricingResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse users.getPricing response")
	}

	return result.CommandResponse.UserGetPricingResult.PricingTypes, nil
}

// GetDomainPricing retrieves pricing for domain operations (register, renew, transfer)
func (c *Client) GetDomainPricing(ctx context.Context, action string) ([]PricingType, error) {
	return c.GetPricing(ctx, "DOMAIN", "", action)
}

// GetSSLPricing retrieves pricing for SSL certificate operations
func (c *Client) GetSSLPricing(ctx context.Context, action string) ([]PricingType, error) {
	return c.GetPricing(ctx, "SSLCERTIFICATE", "", action)
}

// GetWhoisGuardPricing retrieves pricing for WhoisGuard privacy protection
func (c *Client) GetWhoisGuardPricing(ctx context.Context, action string) ([]PricingType, error) {
	return c.GetPricing(ctx, "WHOISGUARD", "", action)
}

// HasSufficientBalance checks if account has sufficient balance for an amount
func (c *Client) HasSufficientBalance(ctx context.Context, requiredAmount float64) (bool, error) {
	balance, err := c.GetUserBalances(ctx)
	if err != nil {
		return false, err
	}

	return balance.AvailableBalance >= requiredAmount, nil
}

// GetTLDByName retrieves TLD information by name
func (c *Client) GetTLDByName(ctx context.Context, tldName string) (*TLD, error) {
	tlds, err := c.GetTLDList(ctx)
	if err != nil {
		return nil, err
	}

	for _, tld := range tlds {
		if tld.Name == tldName {
			return &tld, nil
		}
	}

	return nil, errors.Errorf("TLD '%s' not found", tldName)
}

// IsTLDSupported checks if a TLD is supported for API operations
func (c *Client) IsTLDSupported(ctx context.Context, tldName, operation string) (bool, error) {
	tld, err := c.GetTLDByName(ctx, tldName)
	if err != nil {
		return false, err
	}

	switch operation {
	case "register":
		return tld.IsApiRegisterable, nil
	case "renew":
		return tld.IsApiRenewable, nil
	case "transfer":
		return tld.IsApiTransferable, nil
	default:
		return false, errors.Errorf("unsupported operation: %s", operation)
	}
}