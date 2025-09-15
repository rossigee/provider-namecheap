package namecheap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_RenewDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainName     string
		years          int
		renewXML       string
		getInfoXML     string
		expectedError  string
		expectSuccess  bool
	}{
		{
			name:       "successful domain renewal",
			domainName: "example.com",
			years:      2,
			renewXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainRenewResult DomainName="example.com" DomainID="123" Renew="true" ChargedAmount="18.50" TransactionID="456" OrderID="789"/>
	</CommandResponse>
</ApiResponse>`,
			getInfoXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainGetInfoResult>
			<DomainDetails ID="123" Name="example.com" User="testuser" Created="2024-01-01T00:00:00Z" Expires="2026-01-01T00:00:00Z" IsExpired="false" IsLocked="false" AutoRenew="false" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true"/>
		</DomainGetInfoResult>
	</CommandResponse>
</ApiResponse>`,
			expectSuccess: true,
		},
		{
			name:       "failed domain renewal",
			domainName: "example.com",
			years:      1,
			renewXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainRenewResult DomainName="example.com" DomainID="123" Renew="false"/>
	</CommandResponse>
</ApiResponse>`,
			expectedError: "domain renewal failed",
		},
		{
			name:       "API error response",
			domainName: "example.com",
			years:      1,
			renewXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="ERROR">
	<Errors>
		<Error Number="2030166">Domain not found</Error>
	</Errors>
</ApiResponse>`,
			expectedError: "Namecheap API Error 2030166: Domain not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				assert.Equal(t, "POST", r.Method)

				if callCount == 1 {
					// First call - domain renewal
					assert.Equal(t, "namecheap.domains.renew", r.URL.Query().Get("Command"))
					assert.Equal(t, tt.domainName, r.URL.Query().Get("DomainName"))
					assert.Equal(t, strconv.Itoa(tt.years), r.URL.Query().Get("Years"))

					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(tt.renewXML))
					require.NoError(t, err)
				} else if callCount == 2 && tt.getInfoXML != "" {
					// Second call - get domain info (only for successful renewals)
					assert.Equal(t, "namecheap.domains.getInfo", r.URL.Query().Get("Command"))
					assert.Equal(t, tt.domainName, r.URL.Query().Get("DomainName"))

					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(tt.getInfoXML))
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			client := &Client{
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 5 * time.Second,
				},
				apiUser:  "testuser",
				apiKey:   "testkey",
				username: "testuser",
				clientIP: "127.0.0.1",
			}

			domain, err := client.RenewDomain(context.Background(), tt.domainName, tt.years)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, domain)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, domain)
				assert.Equal(t, tt.domainName, domain.Name)
				assert.Equal(t, 123, domain.ID)
				assert.Equal(t, 2, callCount) // Verify both API calls were made
			}
		})
	}
}

func TestClient_CheckDomainAvailability(t *testing.T) {
	tests := []struct {
		name           string
		domainNames    []string
		responseXML    string
		expectedCount  int
		expectedError  string
	}{
		{
			name:        "single domain available",
			domainNames: []string{"example.com"},
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainCheckResult>
			<DomainCheckResult Domain="example.com" Available="true" ErrorCode="" Description="" IsPremium="false"/>
		</DomainCheckResult>
	</CommandResponse>
</ApiResponse>`,
			expectedCount: 1,
		},
		{
			name:        "multiple domains mixed availability",
			domainNames: []string{"example.com", "google.com", "newdomain.net"},
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainCheckResult>
			<DomainCheckResult Domain="example.com" Available="false" ErrorCode="" Description="Domain taken"/>
			<DomainCheckResult Domain="google.com" Available="false" ErrorCode="" Description="Domain taken"/>
			<DomainCheckResult Domain="newdomain.net" Available="true" ErrorCode="" Description="" IsPremium="false"/>
		</DomainCheckResult>
	</CommandResponse>
</ApiResponse>`,
			expectedCount: 3,
		},
		{
			name:          "empty domain list",
			domainNames:   []string{},
			expectedError: "at least one domain name must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError != "" && len(tt.domainNames) == 0 {
				// Test error case without server
				client := &Client{}
				results, err := client.CheckDomainAvailability(context.Background(), tt.domainNames)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, results)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "namecheap.domains.check", r.URL.Query().Get("Command"))
				assert.Equal(t, strings.Join(tt.domainNames, ","), r.URL.Query().Get("DomainList"))

				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(tt.responseXML))
				require.NoError(t, err)
			}))
			defer server.Close()

			client := &Client{
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 5 * time.Second,
				},
				apiUser:  "testuser",
				apiKey:   "testkey",
				username: "testuser",
				clientIP: "127.0.0.1",
			}

			results, err := client.CheckDomainAvailability(context.Background(), tt.domainNames)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.expectedCount)

				if len(results) > 0 {
					assert.Equal(t, tt.domainNames[0], results[0].Domain)
				}
			}
		})
	}
}

func TestClient_GetDomains(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainGetListResult>
			<Domain ID="123" Name="example.com" User="testuser" Created="2024-01-01T00:00:00Z" Expires="2025-01-01T00:00:00Z" IsExpired="false" IsLocked="false" AutoRenew="false" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true"/>
			<Domain ID="124" Name="test.com" User="testuser" Created="2024-01-01T00:00:00Z" Expires="2025-01-01T00:00:00Z" IsExpired="false" IsLocked="false" AutoRenew="true" WhoisGuard="DISABLED" IsPremium="false" IsOurDNS="false"/>
		</DomainGetListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "namecheap.domains.getList", r.URL.Query().Get("Command"))
		assert.Equal(t, "100", r.URL.Query().Get("PageSize"))

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(responseXML))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		apiUser:  "testuser",
		apiKey:   "testkey",
		username: "testuser",
		clientIP: "127.0.0.1",
	}

	domains, err := client.GetDomains(context.Background())

	assert.NoError(t, err)
	assert.Len(t, domains, 2)
	assert.Equal(t, "example.com", domains[0].Name)
	assert.Equal(t, "test.com", domains[1].Name)
	assert.Equal(t, 123, domains[0].ID)
	assert.Equal(t, 124, domains[1].ID)
}

func TestClient_CreateDomain(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainCreateResult Domain="newdomain.com" Registered="true" ChargedAmount="12.50" DomainID="125" OrderID="456" TransactionID="789" WhoisguardEnable="false" NonRealTimeDomain="false"/>
	</CommandResponse>
</ApiResponse>`

	getInfoXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainGetInfoResult>
			<DomainDetails ID="125" Name="newdomain.com" User="testuser" Created="2024-01-01T00:00:00Z" Expires="2025-01-01T00:00:00Z" IsExpired="false" IsLocked="false" AutoRenew="false" WhoisGuard="DISABLED" IsPremium="false" IsOurDNS="true"/>
		</DomainGetInfoResult>
	</CommandResponse>
</ApiResponse>`

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		assert.Equal(t, "POST", r.Method)

		if callCount == 1 {
			// First call - domain creation
			assert.Equal(t, "namecheap.domains.create", r.URL.Query().Get("Command"))
			assert.Equal(t, "newdomain.com", r.URL.Query().Get("DomainName"))
			assert.Equal(t, "2", r.URL.Query().Get("Years"))

			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(responseXML))
			require.NoError(t, err)
		} else {
			// Second call - get domain info
			assert.Equal(t, "namecheap.domains.getInfo", r.URL.Query().Get("Command"))
			assert.Equal(t, "newdomain.com", r.URL.Query().Get("DomainName"))

			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(getInfoXML))
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		apiUser:  "testuser",
		apiKey:   "testkey",
		username: "testuser",
		clientIP: "127.0.0.1",
	}

	domain, err := client.CreateDomain(context.Background(), "newdomain.com", 2)

	assert.NoError(t, err)
	assert.NotNil(t, domain)
	assert.Equal(t, "newdomain.com", domain.Name)
	assert.Equal(t, 125, domain.ID)
	assert.Equal(t, 2, callCount) // Verify both API calls were made
}