package namecheap

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// SSLCertificate represents an SSL certificate
type SSLCertificate struct {
	CertificateID   int       `xml:"CertificateID,attr"`
	HostName        string    `xml:"HostName,attr"`
	SSLType         string    `xml:"SSLType,attr"`
	PurchaseDate    time.Time `xml:"PurchaseDate,attr"`
	ExpireDate      time.Time `xml:"ExpireDate,attr"`
	ActivationExpireDate time.Time `xml:"ActivationExpireDate,attr"`
	IsExpiredYN     bool      `xml:"IsExpiredYN,attr"`
	Status          string    `xml:"Status,attr"`
	StatusDescription string  `xml:"StatusDescription,attr"`
	Years           int       `xml:"Years,attr"`
}

// SSLListResponse represents the response from ssl.getList
type SSLListResponse struct {
	APIResponse
	CommandResponse struct {
		SSLGetListResult struct {
			SSLCertificates []SSLCertificate `xml:"SSL"`
		} `xml:"SSLGetListResult"`
	} `xml:"CommandResponse"`
}

// SSLCreateResponse represents the response from ssl.create
type SSLCreateResponse struct {
	APIResponse
	CommandResponse struct {
		SSLCreateResult struct {
			IsSuccess     bool    `xml:"IsSuccess,attr"`
			OrderID       int     `xml:"OrderID,attr"`
			TransactionID int     `xml:"TransactionID,attr"`
			ChargedAmount float64 `xml:"ChargedAmount,attr"`
			SSLCertificateID int  `xml:"SSLCertificateID,attr"`
		} `xml:"SSLCreateResult"`
	} `xml:"CommandResponse"`
}

// SSLActivateResponse represents the response from ssl.activate
type SSLActivateResponse struct {
	APIResponse
	CommandResponse struct {
		SSLActivateResult struct {
			IsSuccess bool   `xml:"IsSuccess,attr"`
			ID        int    `xml:"ID,attr"`
		} `xml:"SSLActivateResult"`
	} `xml:"CommandResponse"`
}

// SSLGetInfoResponse represents the response from ssl.getInfo
type SSLGetInfoResponse struct {
	APIResponse
	CommandResponse struct {
		SSLGetInfoResult struct {
			CertificateID        int       `xml:"CertificateID,attr"`
			HostName             string    `xml:"HostName,attr"`
			SSLType              string    `xml:"SSLType,attr"`
			PurchaseDate         time.Time `xml:"PurchaseDate,attr"`
			ExpireDate           time.Time `xml:"ExpireDate,attr"`
			ActivationExpireDate time.Time `xml:"ActivationExpireDate,attr"`
			IsExpiredYN          bool      `xml:"IsExpiredYN,attr"`
			Status               string    `xml:"Status,attr"`
			StatusDescription    string    `xml:"StatusDescription,attr"`
			Years                int       `xml:"Years,attr"`
			Provider             struct {
				Name            string `xml:"Name,attr"`
				DisplayName     string `xml:"DisplayName,attr"`
				LogoURL         string `xml:"LogoURL,attr"`
			} `xml:"Provider"`
			ApproverEmailList    []string `xml:"ApproverEmailList>Email"`
		} `xml:"SSLGetInfoResult"`
	} `xml:"CommandResponse"`
}

// SSLResendResponse represents the response from ssl.resend
type SSLResendResponse struct {
	APIResponse
	CommandResponse struct {
		SSLResendResult struct {
			IsSuccess bool `xml:"IsSuccess,attr"`
		} `xml:"SSLResendResult"`
	} `xml:"CommandResponse"`
}

// SSLReissueResponse represents the response from ssl.reissue
type SSLReissueResponse struct {
	APIResponse
	CommandResponse struct {
		SSLReissueResult struct {
			IsSuccess bool `xml:"IsSuccess,attr"`
		} `xml:"SSLReissueResult"`
	} `xml:"CommandResponse"`
}

// GetSSLCertificates retrieves all SSL certificates for the account
func (c *Client) GetSSLCertificates(ctx context.Context) ([]SSLCertificate, error) {
	resp, err := c.makeRequest(ctx, "namecheap.ssl.getList", map[string]string{
		"PageSize": "100",
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make ssl.getList request")
	}

	var result SSLListResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse ssl.getList response")
	}

	return result.CommandResponse.SSLGetListResult.SSLCertificates, nil
}

// CreateSSLCertificate purchases a new SSL certificate
func (c *Client) CreateSSLCertificate(ctx context.Context, certificateType, years int, sansToAdd string) (int, error) {
	params := map[string]string{
		"Type":  strconv.Itoa(certificateType),
		"Years": strconv.Itoa(years),
	}

	if sansToAdd != "" {
		params["SANStoAdd"] = sansToAdd
	}

	resp, err := c.makeRequest(ctx, "namecheap.ssl.create", params)
	if err != nil {
		return 0, errors.Wrap(err, "failed to make ssl.create request")
	}

	var result SSLCreateResponse
	if err := parseResponse(resp, &result); err != nil {
		return 0, errors.Wrap(err, "failed to parse ssl.create response")
	}

	if !result.CommandResponse.SSLCreateResult.IsSuccess {
		return 0, errors.New("SSL certificate creation failed")
	}

	return result.CommandResponse.SSLCreateResult.SSLCertificateID, nil
}

// ActivateSSLCertificate activates an SSL certificate
func (c *Client) ActivateSSLCertificate(ctx context.Context, certificateID int, csr, domainName, approverEmail, httpDCValidation, dnsValidation, webServerType string) error {
	params := map[string]string{
		"CertificateID": strconv.Itoa(certificateID),
		"CSR":           csr,
		"DomainName":    domainName,
		"ApproverEmail": approverEmail,
	}

	if httpDCValidation != "" {
		params["HTTPDCValidation"] = httpDCValidation
	}

	if dnsValidation != "" {
		params["DNSValidation"] = dnsValidation
	}

	if webServerType != "" {
		params["WebServerType"] = webServerType
	}

	resp, err := c.makeRequest(ctx, "namecheap.ssl.activate", params)
	if err != nil {
		return errors.Wrap(err, "failed to make ssl.activate request")
	}

	var result SSLActivateResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse ssl.activate response")
	}

	if !result.CommandResponse.SSLActivateResult.IsSuccess {
		return errors.New("SSL certificate activation failed")
	}

	return nil
}

// GetSSLCertificate retrieves detailed information about a specific SSL certificate
func (c *Client) GetSSLCertificate(ctx context.Context, certificateID int) (*SSLGetInfoResponse, error) {
	params := map[string]string{
		"CertificateID": strconv.Itoa(certificateID),
	}

	resp, err := c.makeRequest(ctx, "namecheap.ssl.getInfo", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make ssl.getInfo request")
	}

	var result SSLGetInfoResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse ssl.getInfo response")
	}

	return &result, nil
}

// ResendSSLApprovalEmail resends the SSL certificate approval email
func (c *Client) ResendSSLApprovalEmail(ctx context.Context, certificateID int) error {
	params := map[string]string{
		"CertificateID": strconv.Itoa(certificateID),
	}

	resp, err := c.makeRequest(ctx, "namecheap.ssl.resend", params)
	if err != nil {
		return errors.Wrap(err, "failed to make ssl.resend request")
	}

	var result SSLResendResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse ssl.resend response")
	}

	if !result.CommandResponse.SSLResendResult.IsSuccess {
		return errors.New("failed to resend SSL approval email")
	}

	return nil
}

// ReissueSSLCertificate reissues an SSL certificate
func (c *Client) ReissueSSLCertificate(ctx context.Context, certificateID int, csr, approverEmail string) error {
	params := map[string]string{
		"CertificateID": strconv.Itoa(certificateID),
		"CSR":           csr,
		"ApproverEmail": approverEmail,
	}

	resp, err := c.makeRequest(ctx, "namecheap.ssl.reissue", params)
	if err != nil {
		return errors.Wrap(err, "failed to make ssl.reissue request")
	}

	var result SSLReissueResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse ssl.reissue response")
	}

	if !result.CommandResponse.SSLReissueResult.IsSuccess {
		return errors.New("SSL certificate reissue failed")
	}

	return nil
}

// GetSSLCertificatesByDomain retrieves SSL certificates for a specific domain
func (c *Client) GetSSLCertificatesByDomain(ctx context.Context, domainName string) ([]SSLCertificate, error) {
	certificates, err := c.GetSSLCertificates(ctx)
	if err != nil {
		return nil, err
	}

	var domainCertificates []SSLCertificate
	for _, cert := range certificates {
		if strings.EqualFold(cert.HostName, domainName) ||
		   strings.HasSuffix(strings.ToLower(cert.HostName), "."+strings.ToLower(domainName)) {
			domainCertificates = append(domainCertificates, cert)
		}
	}

	return domainCertificates, nil
}

// SSLCertificateExists checks if an SSL certificate exists for a domain
func (c *Client) SSLCertificateExists(ctx context.Context, domainName string) (bool, error) {
	certificates, err := c.GetSSLCertificatesByDomain(ctx, domainName)
	if err != nil {
		return false, err
	}

	return len(certificates) > 0, nil
}