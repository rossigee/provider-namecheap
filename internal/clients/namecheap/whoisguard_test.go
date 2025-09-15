package namecheap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetWhoisGuards(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardGetListResult>
			<Whoisguard ID="123" DomainName="example.com" Created="2024-01-01T00:00:00Z" Status="ENABLED">
				<EmailDetails ForwardedTo="user@email.com" LastAutoEmailDate="2024-01-01T12:00:00Z" AutoEmailCount="5"/>
			</Whoisguard>
			<Whoisguard ID="124" DomainName="test.com" Created="2024-01-01T00:00:00Z" Status="DISABLED">
				<EmailDetails ForwardedTo="" LastAutoEmailDate="" AutoEmailCount="0"/>
			</Whoisguard>
		</WhoisguardGetListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "namecheap.whoisguard.getList", r.URL.Query().Get("Command"))

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

	whoisGuards, err := client.GetWhoisGuards(context.Background())

	assert.NoError(t, err)
	assert.Len(t, whoisGuards, 2)

	// Check first WhoisGuard (enabled)
	assert.Equal(t, 123, whoisGuards[0].ID)
	assert.Equal(t, "example.com", whoisGuards[0].DomainName)
	assert.Equal(t, "ENABLED", whoisGuards[0].Status)
	assert.Equal(t, "user@email.com", whoisGuards[0].EmailDetails.ForwardedTo)

	// Check second WhoisGuard (disabled)
	assert.Equal(t, 124, whoisGuards[1].ID)
	assert.Equal(t, "test.com", whoisGuards[1].DomainName)
	assert.Equal(t, "DISABLED", whoisGuards[1].Status)
}

func TestClient_EnableWhoisGuard(t *testing.T) {
	tests := []struct {
		name           string
		whoisGuardID   int
		domainName     string
		forwardEmail   string
		responseXML    string
		expectedError  string
	}{
		{
			name:         "successful enable",
			whoisGuardID: 123,
			domainName:   "example.com",
			forwardEmail: "user@email.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardEnableResult Domain="example.com" IsSuccess="true"/>
	</CommandResponse>
</ApiResponse>`,
		},
		{
			name:         "enable without forward email",
			whoisGuardID: 123,
			domainName:   "example.com",
			forwardEmail: "",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardEnableResult Domain="example.com" IsSuccess="true"/>
	</CommandResponse>
</ApiResponse>`,
		},
		{
			name:         "failed enable",
			whoisGuardID: 123,
			domainName:   "example.com",
			forwardEmail: "user@email.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardEnableResult Domain="example.com" IsSuccess="false"/>
	</CommandResponse>
</ApiResponse>`,
			expectedError: "failed to enable WhoisGuard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "namecheap.whoisguard.enable", r.URL.Query().Get("Command"))
				assert.Equal(t, "123", r.URL.Query().Get("WhoisguardID"))
				assert.Equal(t, tt.domainName, r.URL.Query().Get("DomainName"))

				if tt.forwardEmail != "" {
					assert.Equal(t, tt.forwardEmail, r.URL.Query().Get("ForwardedToEmail"))
				}

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

			err := client.EnableWhoisGuard(context.Background(), tt.whoisGuardID, tt.domainName, tt.forwardEmail)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_DisableWhoisGuard(t *testing.T) {
	tests := []struct {
		name           string
		whoisGuardID   int
		domainName     string
		responseXML    string
		expectedError  string
	}{
		{
			name:         "successful disable",
			whoisGuardID: 123,
			domainName:   "example.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardDisableResult Domain="example.com" IsSuccess="true"/>
	</CommandResponse>
</ApiResponse>`,
		},
		{
			name:         "failed disable",
			whoisGuardID: 123,
			domainName:   "example.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardDisableResult Domain="example.com" IsSuccess="false"/>
	</CommandResponse>
</ApiResponse>`,
			expectedError: "failed to disable WhoisGuard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "namecheap.whoisguard.disable", r.URL.Query().Get("Command"))
				assert.Equal(t, "123", r.URL.Query().Get("WhoisguardID"))
				assert.Equal(t, tt.domainName, r.URL.Query().Get("DomainName"))

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

			err := client.DisableWhoisGuard(context.Background(), tt.whoisGuardID, tt.domainName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_GetWhoisGuardForDomain(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardGetListResult>
			<Whoisguard ID="123" DomainName="example.com" Created="2024-01-01T00:00:00Z" Status="ENABLED">
				<EmailDetails ForwardedTo="user@email.com" LastAutoEmailDate="2024-01-01T12:00:00Z" AutoEmailCount="5"/>
			</Whoisguard>
			<Whoisguard ID="124" DomainName="test.com" Created="2024-01-01T00:00:00Z" Status="DISABLED">
				<EmailDetails ForwardedTo="" LastAutoEmailDate="" AutoEmailCount="0"/>
			</Whoisguard>
		</WhoisguardGetListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Test finding existing domain
	whoisGuard, err := client.GetWhoisGuardForDomain(context.Background(), "example.com")
	assert.NoError(t, err)
	assert.NotNil(t, whoisGuard)
	assert.Equal(t, 123, whoisGuard.ID)
	assert.Equal(t, "example.com", whoisGuard.DomainName)
	assert.Equal(t, "ENABLED", whoisGuard.Status)

	// Test case insensitive matching
	whoisGuard, err = client.GetWhoisGuardForDomain(context.Background(), "EXAMPLE.COM")
	assert.NoError(t, err)
	assert.NotNil(t, whoisGuard)
	assert.Equal(t, 123, whoisGuard.ID)

	// Test domain not found
	whoisGuard, err = client.GetWhoisGuardForDomain(context.Background(), "notfound.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "WhoisGuard not found for domain")
	assert.Nil(t, whoisGuard)
}

func TestClient_IsWhoisGuardEnabled(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<WhoisguardGetListResult>
			<Whoisguard ID="123" DomainName="example.com" Created="2024-01-01T00:00:00" Status="ENABLED">
				<EmailDetails ForwardedTo="user@email.com"/>
			</Whoisguard>
			<Whoisguard ID="124" DomainName="test.com" Created="2024-01-01T00:00:00" Status="DISABLED">
				<EmailDetails ForwardedTo=""/>
			</Whoisguard>
		</WhoisguardGetListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Test enabled domain
	enabled, err := client.IsWhoisGuardEnabled(context.Background(), "example.com")
	assert.NoError(t, err)
	assert.True(t, enabled)

	// Test disabled domain
	enabled, err = client.IsWhoisGuardEnabled(context.Background(), "test.com")
	assert.NoError(t, err)
	assert.False(t, enabled)

	// Test domain not found
	enabled, err = client.IsWhoisGuardEnabled(context.Background(), "notfound.com")
	assert.NoError(t, err)
	assert.False(t, enabled)
}