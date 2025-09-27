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

func TestClient_GetSSLCertificates(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLGetListResult>
			<SSL CertificateID="123" HostName="example.com" SSLType="PositiveSSL" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" Status="ACTIVE" StatusDescription="Certificate is active" Years="1"/>
			<SSL CertificateID="124" HostName="test.com" SSLType="EssentialSSL" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" Status="PENDING" StatusDescription="Certificate is pending activation" Years="1"/>
		</SSLGetListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "namecheap.ssl.getList", r.URL.Query().Get("Command"))
		assert.Equal(t, "100", r.URL.Query().Get("PageSize"))

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(responseXML))
		require.NoError(t, err)
	}))
	defer server.Close()

	config := Config{
		APIUser:  "testuser",
		APIKey:   "testkey",
		Username: "testuser",
		ClientIP: "127.0.0.1",
		BaseURL:  server.URL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	client := NewClient(config)

	certificates, err := client.GetSSLCertificates(context.Background())

	assert.NoError(t, err)
	assert.Len(t, certificates, 2)

	// Check first certificate
	assert.Equal(t, 123, certificates[0].CertificateID)
	assert.Equal(t, "example.com", certificates[0].HostName)
	assert.Equal(t, "PositiveSSL", certificates[0].SSLType)
	assert.Equal(t, "ACTIVE", certificates[0].Status)
	assert.False(t, certificates[0].IsExpiredYN)

	// Check second certificate
	assert.Equal(t, 124, certificates[1].CertificateID)
	assert.Equal(t, "test.com", certificates[1].HostName)
	assert.Equal(t, "EssentialSSL", certificates[1].SSLType)
	assert.Equal(t, "PENDING", certificates[1].Status)
}

func TestClient_CreateSSLCertificate(t *testing.T) {
	tests := []struct {
		name            string
		certificateType int
		years           int
		sansToAdd       string
		responseXML     string
		expectedCertID  int
		expectedError   string
	}{
		{
			name:            "successful creation",
			certificateType: 1,
			years:           1,
			sansToAdd:       "",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLCreateResult IsSuccess="true" OrderID="456" TransactionID="789" ChargedAmount="12.50" SSLCertificateID="123"/>
	</CommandResponse>
</ApiResponse>`,
			expectedCertID: 123,
		},
		{
			name:            "successful creation with SANs",
			certificateType: 2,
			years:           2,
			sansToAdd:       "www.example.com,mail.example.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLCreateResult IsSuccess="true" OrderID="457" TransactionID="790" ChargedAmount="25.00" SSLCertificateID="124"/>
	</CommandResponse>
</ApiResponse>`,
			expectedCertID: 124,
		},
		{
			name:            "failed creation",
			certificateType: 1,
			years:           1,
			sansToAdd:       "",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLCreateResult IsSuccess="false"/>
	</CommandResponse>
</ApiResponse>`,
			expectedError: "SSL certificate creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "namecheap.ssl.create", r.URL.Query().Get("Command"))
				assert.Equal(t, string(rune(tt.certificateType+'0')), r.URL.Query().Get("Type"))
				assert.Equal(t, string(rune(tt.years+'0')), r.URL.Query().Get("Years"))

				if tt.sansToAdd != "" {
					assert.Equal(t, tt.sansToAdd, r.URL.Query().Get("SANStoAdd"))
				}

				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(tt.responseXML))
				require.NoError(t, err)
			}))
			defer server.Close()

			config := Config{
				APIUser:  "testuser",
				APIKey:   "testkey",
				Username: "testuser",
				ClientIP: "127.0.0.1",
				BaseURL:  server.URL,
				HTTPClient: &http.Client{
					Timeout: 5 * time.Second,
				},
			}
			client := NewClient(config)

			certID, err := client.CreateSSLCertificate(context.Background(), tt.certificateType, tt.years, tt.sansToAdd)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, 0, certID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCertID, certID)
			}
		})
	}
}

func TestClient_ActivateSSLCertificate(t *testing.T) {
	tests := []struct {
		name               string
		certificateID      int
		csr               string
		domainName        string
		approverEmail     string
		httpDCValidation  string
		dnsValidation     string
		webServerType     string
		responseXML       string
		expectedError     string
	}{
		{
			name:          "successful activation",
			certificateID: 123,
			csr:           "-----BEGIN CERTIFICATE REQUEST-----\nMIICZjCCAU4...\n-----END CERTIFICATE REQUEST-----",
			domainName:    "example.com",
			approverEmail: "admin@example.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLActivateResult IsSuccess="true" ID="123"/>
	</CommandResponse>
</ApiResponse>`,
		},
		{
			name:             "activation with DNS validation",
			certificateID:    123,
			csr:             "-----BEGIN CERTIFICATE REQUEST-----\nMIICZjCCAU4...\n-----END CERTIFICATE REQUEST-----",
			domainName:      "example.com",
			approverEmail:   "admin@example.com",
			dnsValidation:   "DNS_CNAME",
			webServerType:   "Apache",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLActivateResult IsSuccess="true" ID="123"/>
	</CommandResponse>
</ApiResponse>`,
		},
		{
			name:          "failed activation",
			certificateID: 123,
			csr:           "invalid-csr",
			domainName:    "example.com",
			approverEmail: "admin@example.com",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLActivateResult IsSuccess="false"/>
	</CommandResponse>
</ApiResponse>`,
			expectedError: "SSL certificate activation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "namecheap.ssl.activate", r.URL.Query().Get("Command"))
				assert.Equal(t, "123", r.URL.Query().Get("CertificateID"))
				assert.Equal(t, tt.csr, r.URL.Query().Get("CSR"))
				assert.Equal(t, tt.domainName, r.URL.Query().Get("DomainName"))
				assert.Equal(t, tt.approverEmail, r.URL.Query().Get("ApproverEmail"))

				if tt.dnsValidation != "" {
					assert.Equal(t, tt.dnsValidation, r.URL.Query().Get("DNSValidation"))
				}
				if tt.webServerType != "" {
					assert.Equal(t, tt.webServerType, r.URL.Query().Get("WebServerType"))
				}

				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(tt.responseXML))
				require.NoError(t, err)
			}))
			defer server.Close()

			config := Config{
				APIUser:  "testuser",
				APIKey:   "testkey",
				Username: "testuser",
				ClientIP: "127.0.0.1",
				BaseURL:  server.URL,
				HTTPClient: &http.Client{
					Timeout: 5 * time.Second,
				},
			}
			client := NewClient(config)

			err := client.ActivateSSLCertificate(context.Background(), tt.certificateID, tt.csr, tt.domainName, tt.approverEmail, tt.httpDCValidation, tt.dnsValidation, tt.webServerType)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_GetSSLCertificate(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLGetInfoResult CertificateID="123" HostName="example.com" SSLType="PositiveSSL" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" Status="ACTIVE" StatusDescription="Certificate is active" Years="1">
			<Provider Name="Comodo" DisplayName="Comodo CA Limited" LogoURL="https://example.com/logo.png"/>
			<ApproverEmailList>
				<Email>admin@example.com</Email>
				<Email>webmaster@example.com</Email>
			</ApproverEmailList>
		</SSLGetInfoResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "namecheap.ssl.getInfo", r.URL.Query().Get("Command"))
		assert.Equal(t, "123", r.URL.Query().Get("CertificateID"))

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(responseXML))
		require.NoError(t, err)
	}))
	defer server.Close()

	config := Config{
		APIUser:  "testuser",
		APIKey:   "testkey",
		Username: "testuser",
		ClientIP: "127.0.0.1",
		BaseURL:  server.URL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	client := NewClient(config)

	cert, err := client.GetSSLCertificate(context.Background(), 123)

	assert.NoError(t, err)
	assert.NotNil(t, cert)

	result := cert.CommandResponse.SSLGetInfoResult
	assert.Equal(t, 123, result.CertificateID)
	assert.Equal(t, "example.com", result.HostName)
	assert.Equal(t, "PositiveSSL", result.SSLType)
	assert.Equal(t, "ACTIVE", result.Status)
	assert.False(t, result.IsExpiredYN)
	assert.Equal(t, "Comodo", result.Provider.Name)
	assert.Len(t, result.ApproverEmailList, 2)
	assert.Contains(t, result.ApproverEmailList, "admin@example.com")
	assert.Contains(t, result.ApproverEmailList, "webmaster@example.com")
}

func TestClient_GetSSLCertificatesByDomain(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLGetListResult>
			<SSL CertificateID="123" HostName="example.com" SSLType="PositiveSSL" Status="ACTIVE" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" StatusDescription="Certificate is active" Years="1"/>
			<SSL CertificateID="124" HostName="www.example.com" SSLType="EssentialSSL" Status="ACTIVE" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" StatusDescription="Certificate is active" Years="1"/>
			<SSL CertificateID="125" HostName="test.com" SSLType="PositiveSSL" Status="ACTIVE" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" StatusDescription="Certificate is active" Years="1"/>
			<SSL CertificateID="126" HostName="mail.example.com" SSLType="WildcardSSL" Status="PENDING" PurchaseDate="2024-01-01T00:00:00Z" ExpireDate="2025-01-01T00:00:00Z" ActivationExpireDate="2024-12-01T00:00:00Z" IsExpiredYN="false" StatusDescription="Certificate is pending" Years="1"/>
		</SSLGetListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(responseXML))
		require.NoError(t, err)
	}))
	defer server.Close()

	config := Config{
		APIUser:  "testuser",
		APIKey:   "testkey",
		Username: "testuser",
		ClientIP: "127.0.0.1",
		BaseURL:  server.URL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	client := NewClient(config)

	// Test finding certificates for exact domain match
	certs, err := client.GetSSLCertificatesByDomain(context.Background(), "example.com")
	assert.NoError(t, err)
	assert.Len(t, certs, 3) // example.com, www.example.com, mail.example.com

	// Verify the certificates returned
	certIDs := make([]int, len(certs))
	for i, cert := range certs {
		certIDs[i] = cert.CertificateID
	}
	assert.Contains(t, certIDs, 123) // example.com
	assert.Contains(t, certIDs, 124) // www.example.com
	assert.Contains(t, certIDs, 126) // mail.example.com

	// Test finding certificates for different domain
	certs, err = client.GetSSLCertificatesByDomain(context.Background(), "test.com")
	assert.NoError(t, err)
	assert.Len(t, certs, 1)
	assert.Equal(t, 125, certs[0].CertificateID)

	// Test domain with no certificates
	certs, err = client.GetSSLCertificatesByDomain(context.Background(), "notfound.com")
	assert.NoError(t, err)
	assert.Len(t, certs, 0)
}

func TestClient_ResendSSLApprovalEmail(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<SSLResendResult IsSuccess="true"/>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "namecheap.ssl.resend", r.URL.Query().Get("Command"))
		assert.Equal(t, "123", r.URL.Query().Get("CertificateID"))

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(responseXML))
		require.NoError(t, err)
	}))
	defer server.Close()

	config := Config{
		APIUser:  "testuser",
		APIKey:   "testkey",
		Username: "testuser",
		ClientIP: "127.0.0.1",
		BaseURL:  server.URL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	client := NewClient(config)

	err := client.ResendSSLApprovalEmail(context.Background(), 123)
	assert.NoError(t, err)
}