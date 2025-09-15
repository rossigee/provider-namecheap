# provider-namecheap

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-namecheap/ci.yml?branch=master)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/rossigee/provider-namecheap)
[![Version](https://img.shields.io/github/v/release/rossigee/provider-namecheap)][releases]
[![GitHub downloads](https://img.shields.io/github/downloads/rossigee/provider-namecheap/total)][releases]

[build]: https://github.com/rossigee/provider-namecheap/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-namecheap/releases

**✅ BUILD STATUS: WORKING** - Successfully builds and passes all tests (v0.1.0)

Crossplane provider for managing Namecheap domains and DNS records with full v2 support.

## Overview

This provider enables you to manage Namecheap resources declaratively using Kubernetes manifests. It implements Crossplane v2 patterns with namespaced resources for better multi-tenancy support.

## Features

- **Domain Management**: Registration, renewals, and nameserver configuration
- **DNS Record Management**: Full CRUD operations for A, AAAA, CNAME, MX, TXT, SRV records
- **Crossplane v2 Support**: Namespaced resources with `.m.` API groups for multi-tenancy
- **Sandbox Mode**: Test without real charges using Namecheap's sandbox environment
- **Provider Status**: ✅ Production ready with standardized CI/CD pipeline

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-namecheap:v0.1.0`
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
  package: ghcr.io/rossigee/provider-namecheap:v0.1.0
EOF
```

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

## Configuration

### ProviderConfig Options

- `credentials` - API credentials configuration
- `apiBase` - Custom API base URL (optional)
- `sandboxMode` - Enable sandbox mode for testing (default: false)

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
- `go` (1.24+)
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

### Debug Mode

Enable debug logging by setting the debug flag on the provider deployment:

```bash
kubectl patch deployment provider-namecheap -n crossplane-system \
  --patch '{"spec":{"template":{"spec":{"containers":[{"name":"package-runtime","args":["--debug"]}]}}}}'
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