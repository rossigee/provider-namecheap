# provider-namecheap

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-namecheap/ci.yml?branch=master)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/rossigee/provider-namecheap)
[![Version](https://img.shields.io/github/v/release/rossigee/provider-namecheap)][releases]
[![GitHub downloads](https://img.shields.io/github/downloads/rossigee/provider-namecheap/total)][releases]

[build]: https://github.com/rossigee/provider-namecheap/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-namecheap/releases

**✅ BUILD STATUS: WORKING** - Successfully builds and passes all tests (v0.3.2)

Crossplane provider for comprehensive Namecheap service management with full v2 support and extensive API coverage.

## Overview

This provider enables you to manage Namecheap resources declaratively using Kubernetes manifests. It implements Crossplane v2 patterns with namespaced resources for better multi-tenancy support and covers the complete Namecheap API surface.

## Features

- **Domain Management**: Registration, renewals, transfers, availability checking, and nameserver configuration
- **DNS Record Management**: Full CRUD operations for A, AAAA, CNAME, MX, TXT, SRV records with batch operations
- **SSL Certificate Management**: Complete lifecycle management including purchase, activation, renewal, and reissue
- **WhoisGuard Privacy Protection**: Enable/disable privacy protection services for domains
- **Account Management**: Balance checking, pricing retrieval, TLD support verification
- **Crossplane v2 Support**: Namespaced resources with `.m.` API groups for multi-tenancy
- **Sandbox Mode**: Test without real charges using Namecheap's sandbox environment
- **Comprehensive Testing**: 51.0% test coverage with 22 test functions (42 test executions) across all APIs
- **Provider Status**: ✅ Production ready with standardized CI/CD pipeline

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-namecheap:v0.3.2`
- **Harbor**: Available via environment configuration
- **Upbound**: Available via environment configuration

## Quick Start

### Prerequisites

- Kubernetes cluster with [Crossplane](https://crossplane.io/) v1.18.0+ installed
- Namecheap account with API access enabled
- Your Namecheap API credentials

### Installation

1. **Install the provider:**

```bash
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-namecheap
spec:
  package: ghcr.io/rossigee/provider-namecheap:v0.3.2
EOF
```

### Available Resources

| Resource | API Version | Scope | Description |
|----------|-------------|-------|-------------|
| `Domain` | `namecheap.m.crossplane.io/v1beta1` | Namespaced | Domain registration and management |
| `DNSRecord` | `namecheap.m.crossplane.io/v1beta1` | Namespaced | DNS record management |
| `SSLCertificate` | `namecheap.m.crossplane.io/v1beta1` | Namespaced | SSL certificate lifecycle management |
| `ProviderConfig` | `namecheap.m.crossplane.io/v1beta1` | Namespaced | Provider configuration |

2. **Create a secret with your Namecheap API credentials:**

```bash
kubectl create secret generic namecheap-creds -n crossplane-system \
  --from-literal=credentials='{
    "apiUser": "your-api-user",
    "apiKey": "your-api-key",
    "username": "your-username",
    "clientIP": "your-client-ip"
  }'
```

3. **Create a ProviderConfig:**

```yaml
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: namecheap-creds
      key: credentials
  sandboxMode: true  # Set to false for production
```

### Usage Examples

#### Domain Management

```yaml
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: example-domain
  namespace: production
spec:
  forProvider:
    domainName: example.com
    registrationYears: 1
    nameservers:
      - ns1.example.com
      - ns2.example.com
  providerConfigRef:
    name: default
  deletionPolicy: Delete
```

#### DNS Record Management

```yaml
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: DNSRecord
metadata:
  name: www-record
  namespace: production
spec:
  forProvider:
    domain: example.com
    name: www
    type: A
    value: 192.168.1.100
    ttl: 300
  providerConfigRef:
    name: default
  deletionPolicy: Delete
```

## API Reference

### Domain

The `Domain` resource manages domain registration and configuration.

**Spec Fields:**
- `domainName` (string, required) - The domain name to register/manage
- `registrationYears` (int, optional) - Years to register domain (default: 1)
- `nameservers` ([]string, optional) - Custom nameservers for the domain

**Status Fields:**
- `id` (string) - Namecheap domain ID
- `status` (string) - Domain status
- `createdDate` (timestamp) - Domain creation date
- `expirationDate` (timestamp) - Domain expiration date

### DNSRecord

The `DNSRecord` resource manages DNS records for domains.

**Spec Fields:**
- `domain` (string, required) - The domain name
- `name` (string, required) - Record name (e.g., "www", "@")
- `type` (string, required) - Record type: A, AAAA, CNAME, MX, TXT, SRV
- `value` (string, required) - Record value
- `ttl` (int, optional) - Time to live in seconds (default: 300)
- `priority` (int, optional) - Priority for MX/SRV records

**Status Fields:**
- `id` (string) - Namecheap record ID
- `fqdn` (string) - Fully qualified domain name

### SSLCertificate

The `SSLCertificate` resource manages SSL certificate lifecycle including purchase, activation, and renewal.

**Spec Fields:**
- `certificateType` (int, required) - SSL certificate type ID from Namecheap
- `domainName` (string, required) - Primary domain for the certificate
- `years` (int, optional) - Certificate validity period (1-3 years, default: 1)
- `sansToAdd` (string, optional) - Additional Subject Alternative Names
- `csr` (string, optional) - Certificate Signing Request for activation
- `approverEmail` (string, optional) - Email for certificate approval
- `autoActivate` (bool, optional) - Automatically activate after purchase
- `httpDCValidation` (string, optional) - HTTP domain control validation
- `dnsValidation` (string, optional) - DNS domain control validation
- `webServerType` (string, optional) - Web server type (apache, iis, nginx, etc.)

**Status Fields:**
- `certificateID` (int) - Namecheap certificate ID
- `hostName` (string) - Certificate hostname
- `sslType` (string) - SSL certificate type name
- `status` (string) - Certificate status (ACTIVE, PENDING, etc.)
- `purchaseDate` (timestamp) - Certificate purchase date
- `expireDate` (timestamp) - Certificate expiration date
- `activationExpireDate` (timestamp) - Activation deadline
- `providerName` (string) - SSL provider name
- `approverEmailList` ([]string) - Valid approver email addresses

#### SSL Certificate Management

```yaml
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: SSLCertificate
metadata:
  name: example-ssl-cert
  namespace: production
spec:
  forProvider:
    certificateType: 1  # Basic SSL certificate type
    domainName: example.com
    years: 1
    autoActivate: true
    csr: |
      -----BEGIN CERTIFICATE REQUEST-----
      ...your CSR content...
      -----END CERTIFICATE REQUEST-----
    approverEmail: admin@example.com
    webServerType: apache
  providerConfigRef:
    name: default
  deletionPolicy: Delete
```

#### WhoisGuard Privacy Protection

```yaml
# Enable WhoisGuard privacy protection
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: example-domain-private
  namespace: production
spec:
  forProvider:
    domainName: example.com
    registrationYears: 1
    privacyProtection: true  # Enable WhoisGuard privacy
  providerConfigRef:
    name: default
  deletionPolicy: Delete
```

### Advanced SSL Certificate Operations

SSL certificates support additional operations via annotations:

```yaml
# Reissue an existing certificate
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: SSLCertificate
metadata:
  name: example-ssl-cert
  namespace: production
  annotations:
    namecheap.crossplane.io/reissue: "true"  # Trigger reissue
spec:
  forProvider:
    certificateType: 1
    domainName: example.com
    csr: |
      -----BEGIN CERTIFICATE REQUEST-----
      ...new CSR content...
      -----END CERTIFICATE REQUEST-----
    approverEmail: admin@example.com
  providerConfigRef:
    name: default
---
# Resend approval email
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: SSLCertificate
metadata:
  name: example-ssl-cert-approval
  namespace: production
  annotations:
    namecheap.crossplane.io/resend-approval: "true"  # Resend approval email
spec:
  # ... certificate spec
```

## Configuration

### ProviderConfig Complete Example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: namecheap-credentials
  namespace: crossplane-system
type: Opaque
data:
  credentials: ewogICJhcGlVc2VyIjogInlvdXJfYXBpX3VzZXIiLAogICJhcGlLZXkiOiAieW91cl9hcGlfa2V5IiwKICAidXNlcm5hbWUiOiAieW91cl91c2VybmFtZSIsCiAgImNsaWVudElQIjogInlvdXJfY2xpZW50X2lwIgp9
---
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: namecheap-credentials
      namespace: crossplane-system
      key: credentials
  sandboxMode: true  # Set to false for production
```

### ProviderConfig Options

- `credentials` - API credentials configuration (JSON format)
- `sandboxMode` - Enable sandbox mode for testing (default: false)

### Credentials JSON Format

```json
{
  "apiUser": "your_api_user",
  "apiKey": "your_api_key",
  "username": "your_username",
  "clientIP": "your_client_ip"
}
```

### Namecheap API Setup

1. **Enable API Access:**
   - Log into your Namecheap account
   - Go to Profile → Tools → Namecheap API Access
   - Enable API access and whitelist your IP address

2. **Get API Credentials:**
   - API User: Your Namecheap username
   - API Key: Generated API key from your account
   - Username: Your Namecheap username
   - Client IP: Your server's public IP address

## Local Development

### Requirements

- `docker`
- `go` (1.25.1+)
- `make`
- `kubectl`
- `git` with submodules
- `pre-commit` (optional, for development with quality gates)

### Common make targets

- `make build` to build the binary and docker image
- `make generate` to (re)generate additional code artifacts
- `make lint` to run linting and code quality checks
- `make test` run test suite
- `make reviewable` for full pre-commit validation
- `make publish` to build, package, and publish to registry

### Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Troubleshooting

### Common Issues

**Provider won't start:**
- Check that your IP address is whitelisted in Namecheap API settings
- Verify API credentials are correct in the secret
- Check provider logs: `kubectl logs -n crossplane-system deployment/provider-namecheap`

**DNS records not updating:**
- Ensure domain uses Namecheap's DNS hosting
- Check TTL values - changes may take time to propagate
- Verify record syntax is correct for the record type

**Domain operations failing:**
- Check if domain is actually registered with Namecheap
- Verify your account has sufficient privileges
- For sandbox mode, use test domains provided by Namecheap

**SSL certificate operations failing:**
- Ensure CSR format is valid PEM format with BEGIN/END markers
- Check that domain is verified and accessible
- Verify certificate type is supported by Namecheap
- Check approval email is valid and accessible

**WhoisGuard operations failing:**
- Verify domain is registered with Namecheap
- Check that WhoisGuard is available for the domain TLD
- Ensure account has sufficient balance for WhoisGuard services

### Testing and Validation

**Test your configuration:**
```bash
# Test domain availability
kubectl apply -f - <<EOF
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: test-availability
  namespace: default
spec:
  forProvider:
    domainName: test-example-$(date +%s).com
    registrationYears: 1
  providerConfigRef:
    name: default
  deletionPolicy: Delete
EOF

# Check resource status
kubectl describe domain test-availability -n default
```

**Validate SSL certificate workflow:**
```bash
# Create test SSL certificate
kubectl apply -f examples/ssl-certificate-basic.yaml

# Monitor certificate status
kubectl get sslcertificate -n default -w
```

### Debug Mode

Enable debug logging by setting the debug flag on the provider deployment:

```bash
kubectl patch deployment provider-namecheap -n crossplane-system \
  --patch '{"spec":{"template":{"spec":{"containers":[{"name":"package-runtime","args":["--debug"]}]}}}}'
```

### Performance and Monitoring

**Provider Health Check:**
```bash
# Check provider status
kubectl get providers.pkg.crossplane.io provider-namecheap

# Monitor provider logs
kubectl logs -n crossplane-system deployment/provider-namecheap -f

# Check resource reconciliation
kubectl get managed -o wide
```

**Resource Status Monitoring:**
```bash
# Check all Namecheap resources
kubectl get domains,dnsrecords,sslcertificates -A

# Monitor specific resource events
kubectl describe domain example-domain -n production
kubectl get events -n production --field-selector involvedObject.name=example-domain
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/rossigee/provider-namecheap/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rossigee/provider-namecheap/discussions)
- **Documentation**: [Crossplane Documentation](https://crossplane.io/docs/)

## Related Projects

- [Crossplane](https://github.com/crossplane/crossplane) - The cloud native control plane
- [Provider AWS](https://github.com/crossplane-contrib/provider-aws) - AWS provider for Crossplane
- [Provider Terraform](https://github.com/crossplane-contrib/provider-terraform) - Terraform provider for Crossplane