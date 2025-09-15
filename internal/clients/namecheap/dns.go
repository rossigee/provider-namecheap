package namecheap

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// DNSRecord represents a DNS record in Namecheap
type DNSRecord struct {
	HostID     int    `xml:"HostId,attr"`
	Name       string `xml:"Name,attr"`
	Type       string `xml:"Type,attr"`
	Address    string `xml:"Address,attr"`
	MXPref     int    `xml:"MXPref,attr"`
	TTL        int    `xml:"TTL,attr"`
	AssociatedAppTitle string `xml:"AssociatedAppTitle,attr"`
	FriendlyName       string `xml:"FriendlyName,attr"`
	IsActive           bool   `xml:"IsActive,attr"`
	IsDDNSEnabled      bool   `xml:"IsDDNSEnabled,attr"`
}

// DNSHostsResponse represents the response from domains.dns.getHosts
type DNSHostsResponse struct {
	APIResponse
	CommandResponse struct {
		DomainDNSGetHostsResult struct {
			Domain    string      `xml:"Domain,attr"`
			IsUsingOurDNS bool    `xml:"IsUsingOurDNS,attr"`
			Hosts     []DNSRecord `xml:"host"`
		} `xml:"DomainDNSGetHostsResult"`
	} `xml:"CommandResponse"`
}

// DNSSetHostsResponse represents the response from domains.dns.setHosts
type DNSSetHostsResponse struct {
	APIResponse
	CommandResponse struct {
		DomainDNSSetHostsResult struct {
			Domain    string `xml:"Domain,attr"`
			IsSuccess bool   `xml:"IsSuccess,attr"`
		} `xml:"DomainDNSSetHostsResult"`
	} `xml:"CommandResponse"`
}

// GetDNSRecords retrieves all DNS records for a domain
func (c *Client) GetDNSRecords(ctx context.Context, domainName string) ([]DNSRecord, error) {
	parts := strings.Split(domainName, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid domain name format")
	}

	params := map[string]string{
		"SLD": parts[0],
		"TLD": strings.Join(parts[1:], "."),
	}

	resp, err := c.makeRequest(ctx, "namecheap.domains.dns.getHosts", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make domains.dns.getHosts request")
	}

	var result DNSHostsResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to parse domains.dns.getHosts response")
	}

	return result.CommandResponse.DomainDNSGetHostsResult.Hosts, nil
}

// GetDNSRecord retrieves a specific DNS record by name and type
func (c *Client) GetDNSRecord(ctx context.Context, domainName, recordName, recordType string) (*DNSRecord, error) {
	records, err := c.GetDNSRecords(ctx, domainName)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.Name == recordName && record.Type == recordType {
			return &record, nil
		}
	}

	return nil, errors.New("DNS record not found")
}

// CreateDNSRecord creates a new DNS record
func (c *Client) CreateDNSRecord(ctx context.Context, domainName string, record DNSRecord) error {
	// Get existing records
	existingRecords, err := c.GetDNSRecords(ctx, domainName)
	if err != nil {
		return errors.Wrap(err, "failed to get existing DNS records")
	}

	// Add the new record
	updatedRecords := append(existingRecords, record)

	return c.setDNSRecords(ctx, domainName, updatedRecords)
}

// UpdateDNSRecord updates an existing DNS record
func (c *Client) UpdateDNSRecord(ctx context.Context, domainName string, record DNSRecord) error {
	// Get existing records
	existingRecords, err := c.GetDNSRecords(ctx, domainName)
	if err != nil {
		return errors.Wrap(err, "failed to get existing DNS records")
	}

	// Find and update the record
	found := false
	for i, existingRecord := range existingRecords {
		if existingRecord.HostID == record.HostID ||
		   (existingRecord.Name == record.Name && existingRecord.Type == record.Type) {
			existingRecords[i] = record
			found = true
			break
		}
	}

	if !found {
		return errors.New("DNS record not found for update")
	}

	return c.setDNSRecords(ctx, domainName, existingRecords)
}

// DeleteDNSRecord deletes a DNS record
func (c *Client) DeleteDNSRecord(ctx context.Context, domainName string, recordName, recordType string) error {
	// Get existing records
	existingRecords, err := c.GetDNSRecords(ctx, domainName)
	if err != nil {
		return errors.Wrap(err, "failed to get existing DNS records")
	}

	// Filter out the record to delete
	var updatedRecords []DNSRecord
	found := false
	for _, record := range existingRecords {
		if record.Name == recordName && record.Type == recordType {
			found = true
			continue // Skip this record (delete it)
		}
		updatedRecords = append(updatedRecords, record)
	}

	if !found {
		return errors.New("DNS record not found for deletion")
	}

	return c.setDNSRecords(ctx, domainName, updatedRecords)
}

// setDNSRecords sets all DNS records for a domain (replaces existing records)
func (c *Client) setDNSRecords(ctx context.Context, domainName string, records []DNSRecord) error {
	parts := strings.Split(domainName, ".")
	if len(parts) < 2 {
		return errors.New("invalid domain name format")
	}

	params := map[string]string{
		"SLD": parts[0],
		"TLD": strings.Join(parts[1:], "."),
	}

	// Add each record as a parameter
	for i, record := range records {
		base := "HostName" + strconv.Itoa(i+1)
		params[base] = record.Name
		params["RecordType"+strconv.Itoa(i+1)] = record.Type
		params["Address"+strconv.Itoa(i+1)] = record.Address

		if record.TTL > 0 {
			params["TTL"+strconv.Itoa(i+1)] = strconv.Itoa(record.TTL)
		}

		if record.Type == "MX" && record.MXPref > 0 {
			params["MXPref"+strconv.Itoa(i+1)] = strconv.Itoa(record.MXPref)
		}
	}

	resp, err := c.makeRequest(ctx, "namecheap.domains.dns.setHosts", params)
	if err != nil {
		return errors.Wrap(err, "failed to make domains.dns.setHosts request")
	}

	var result DNSSetHostsResponse
	if err := parseResponse(resp, &result); err != nil {
		return errors.Wrap(err, "failed to parse domains.dns.setHosts response")
	}

	if !result.CommandResponse.DomainDNSSetHostsResult.IsSuccess {
		return errors.New("failed to update DNS records")
	}

	return nil
}

// DNSRecordExists checks if a DNS record exists
func (c *Client) DNSRecordExists(ctx context.Context, domainName, recordName, recordType string) (bool, error) {
	_, err := c.GetDNSRecord(ctx, domainName, recordName, recordType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}