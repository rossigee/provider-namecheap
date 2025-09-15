package namecheap

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetUserBalances(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<UserGetBalancesResult Currency="USD" AvailableBalance="150.75" AccountBalance="150.75" EarnedAmount="25.00" WithdrawableAmount="125.75" FundsRequiredForAutoRenew="50.00"/>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "namecheap.users.getBalances", r.URL.Query().Get("Command"))

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

	balance, err := client.GetUserBalances(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, "USD", balance.Currency)
	assert.Equal(t, 150.75, balance.AvailableBalance)
	assert.Equal(t, 150.75, balance.AccountBalance)
	assert.Equal(t, 25.00, balance.EarnedAmount)
	assert.Equal(t, 125.75, balance.WithdrawableAmount)
	assert.Equal(t, 50.00, balance.FundsRequiredForAutoRenew)
}

func TestClient_GetTLDList(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainsGetTldListResult>
			<Tld Name="com" NonRealTime="false" MinRegisterYears="1" MaxRegisterYears="10" MinRenewYears="1" MaxRenewYears="10" MinTransferYears="1" MaxTransferYears="1" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true" IsEppRequired="false" IsDisableModContact="false" IsDisableWGAllot="false" IsIncludeInExtendedSearchOnly="false" SequenceNumber="10" Type="GTLD" SubType="" IsSupportsIDN="true" Category="A" SupportsRegistrarLock="true" AddGracePeriodFee="0" WhoisVerification="true" ProviderApiDelete="false" TldState="clientTransferProhibited" SearchGroup="com" Registry="Verisign"/>
			<Tld Name="net" NonRealTime="false" MinRegisterYears="1" MaxRegisterYears="10" MinRenewYears="1" MaxRenewYears="10" MinTransferYears="1" MaxTransferYears="1" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true" IsEppRequired="false" IsDisableModContact="false" IsDisableWGAllot="false" IsIncludeInExtendedSearchOnly="false" SequenceNumber="20" Type="GTLD" SubType="" IsSupportsIDN="true" Category="A" SupportsRegistrarLock="true" AddGracePeriodFee="0" WhoisVerification="true" ProviderApiDelete="false" TldState="clientTransferProhibited" SearchGroup="net" Registry="Verisign"/>
		</DomainsGetTldListResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "namecheap.domains.getTldList", r.URL.Query().Get("Command"))

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

	tlds, err := client.GetTLDList(context.Background())

	assert.NoError(t, err)
	assert.Len(t, tlds, 2)

	// Check .com TLD
	com := tlds[0]
	assert.Equal(t, "com", com.Name)
	assert.False(t, com.NonRealTime)
	assert.Equal(t, 1, com.MinRegisterYears)
	assert.Equal(t, 10, com.MaxRegisterYears)
	assert.True(t, com.IsApiRegisterable)
	assert.True(t, com.IsApiRenewable)
	assert.True(t, com.IsApiTransferable)
	assert.Equal(t, "GTLD", com.Type)
	assert.Equal(t, "Verisign", com.Registry)

	// Check .net TLD
	net := tlds[1]
	assert.Equal(t, "net", net.Name)
	assert.True(t, net.IsApiRegisterable)
	assert.True(t, net.IsApiRenewable)
	assert.True(t, net.IsApiTransferable)
}

func TestClient_GetPricing(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<UserGetPricingResult ProductType="DOMAIN" ProductCategory="" Product="">
			<ProductType>
				<PricingType Name="DOMAIN" Price="12.50" RegularPrice="12.50" YourPrice="12.50" YourPriceRange="" PromoPrice="0.00" Currency="USD" Duration="1" DurationType="YEAR" PricingType="REGISTRATION" AdditionalCost="0.00"/>
				<PricingType Name="DOMAIN" Price="12.50" RegularPrice="12.50" YourPrice="12.50" YourPriceRange="" PromoPrice="0.00" Currency="USD" Duration="1" DurationType="YEAR" PricingType="RENEWAL" AdditionalCost="0.00"/>
			</ProductType>
		</UserGetPricingResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "namecheap.users.getPricing", r.URL.Query().Get("Command"))
		assert.Equal(t, "DOMAIN", r.URL.Query().Get("ProductType"))
		assert.Equal(t, "REGISTER", r.URL.Query().Get("Action"))

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

	pricing, err := client.GetPricing(context.Background(), "DOMAIN", "", "REGISTER")

	assert.NoError(t, err)
	assert.Len(t, pricing, 2)

	// Check registration pricing
	reg := pricing[0]
	assert.Equal(t, "DOMAIN", reg.Name)
	assert.Equal(t, 12.50, reg.Price)
	assert.Equal(t, 12.50, reg.RegularPrice)
	assert.Equal(t, 12.50, reg.YourPrice)
	assert.Equal(t, "USD", reg.Currency)
	assert.Equal(t, 1, reg.Duration)
	assert.Equal(t, "YEAR", reg.DurationType)
	assert.Equal(t, "REGISTRATION", reg.PricingType)

	// Check renewal pricing
	ren := pricing[1]
	assert.Equal(t, "RENEWAL", ren.PricingType)
}

func TestClient_GetDomainPricing(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<UserGetPricingResult ProductType="DOMAIN" ProductCategory="" Product="">
			<ProductType>
				<PricingType Name="DOMAIN" Price="12.50" RegularPrice="12.50" YourPrice="12.50" Currency="USD" Duration="1" DurationType="YEAR" PricingType="REGISTRATION"/>
			</ProductType>
		</UserGetPricingResult>
	</CommandResponse>
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DOMAIN", r.URL.Query().Get("ProductType"))
		assert.Equal(t, "REGISTER", r.URL.Query().Get("Action"))

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

	pricing, err := client.GetDomainPricing(context.Background(), "REGISTER")

	assert.NoError(t, err)
	assert.Len(t, pricing, 1)
	assert.Equal(t, "REGISTRATION", pricing[0].PricingType)
}

func TestClient_HasSufficientBalance(t *testing.T) {
	tests := []struct {
		name           string
		requiredAmount float64
		balance        float64
		expected       bool
	}{
		{
			name:           "sufficient balance",
			requiredAmount: 50.00,
			balance:        150.75,
			expected:       true,
		},
		{
			name:           "insufficient balance",
			requiredAmount: 200.00,
			balance:        150.75,
			expected:       false,
		},
		{
			name:           "exact balance",
			requiredAmount: 150.75,
			balance:        150.75,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<UserGetBalancesResult Currency="USD" AvailableBalance="` + fmt.Sprintf("%.2f", tt.balance) + `" AccountBalance="150.75" EarnedAmount="25.00" WithdrawableAmount="125.75" FundsRequiredForAutoRenew="50.00"/>
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

			sufficient, err := client.HasSufficientBalance(context.Background(), tt.requiredAmount)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, sufficient)
		})
	}
}

func TestClient_GetTLDByName(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainsGetTldListResult>
			<Tld Name="com" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true"/>
			<Tld Name="net" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="false"/>
		</DomainsGetTldListResult>
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

	// Test finding existing TLD
	tld, err := client.GetTLDByName(context.Background(), "com")
	assert.NoError(t, err)
	assert.NotNil(t, tld)
	assert.Equal(t, "com", tld.Name)
	assert.True(t, tld.IsApiRegisterable)

	// Test finding another TLD
	tld, err = client.GetTLDByName(context.Background(), "net")
	assert.NoError(t, err)
	assert.NotNil(t, tld)
	assert.Equal(t, "net", tld.Name)
	assert.False(t, tld.IsApiTransferable)

	// Test TLD not found
	tld, err = client.GetTLDByName(context.Background(), "xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TLD 'xyz' not found")
	assert.Nil(t, tld)
}

func TestClient_IsTLDSupported(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse Status="OK">
	<CommandResponse>
		<DomainsGetTldListResult>
			<Tld Name="com" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true"/>
			<Tld Name="net" IsApiRegisterable="false" IsApiRenewable="true" IsApiTransferable="false"/>
		</DomainsGetTldListResult>
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

	// Test .com supports all operations
	supported, err := client.IsTLDSupported(context.Background(), "com", "register")
	assert.NoError(t, err)
	assert.True(t, supported)

	supported, err = client.IsTLDSupported(context.Background(), "com", "renew")
	assert.NoError(t, err)
	assert.True(t, supported)

	supported, err = client.IsTLDSupported(context.Background(), "com", "transfer")
	assert.NoError(t, err)
	assert.True(t, supported)

	// Test .net has limited support
	supported, err = client.IsTLDSupported(context.Background(), "net", "register")
	assert.NoError(t, err)
	assert.False(t, supported)

	supported, err = client.IsTLDSupported(context.Background(), "net", "renew")
	assert.NoError(t, err)
	assert.True(t, supported)

	supported, err = client.IsTLDSupported(context.Background(), "net", "transfer")
	assert.NoError(t, err)
	assert.False(t, supported)

	// Test invalid operation
	supported, err = client.IsTLDSupported(context.Background(), "com", "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported operation: invalid")
	assert.False(t, supported)
}