# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**provider-namecheap** is a Crossplane v2-only provider for managing Namecheap DNS and domain resources through Kubernetes. This provider implements modern Crossplane v2 patterns with namespaced resources and selective resource activation.

**Key Characteristics:**
- **v2-Only Architecture**: Namespaced resources with `.m.` API groups
- **Managed Resource Definitions (MRDs)**: Selective resource activation
- **DeploymentRuntimeConfig**: Modern runtime configuration
- **Namespace Isolation**: Multi-tenant resource management

**Target Resources:**
- Domain registration and management
- DNS record management (A, AAAA, CNAME, MX, TXT, etc.)
- Nameserver configuration
- SSL certificate management (if supported by Namecheap API)

## Essential Build System Setup

### Prerequisites
- **Go 1.24+**: Required for provider development
- **Docker**: For container image builds
- **Make**: Build orchestration
- **Git Submodules**: Critical for build system

### Critical First Steps
```bash
# MANDATORY - Initialize build submodule (rossigee/build fork)
git submodule add https://github.com/rossigee/build build
git submodule update --init --recursive

# Verify build system works
make lint            # Should complete without "No rule to make target" errors
```

## Build Commands

### Essential Make Targets (v2)
```bash
make generate       # Generate namespaced CRDs, MRDs, and boilerplate code
make build          # Build provider binary with v2 support
make lint           # Lint code with golangci-lint (v2 compliance)
make test           # Run unit tests with coverage
make reviewable     # Full pre-commit validation (generate, lint, test, v2 checks)
make publish        # Build, package, and publish to registry with MRDs
make xpkg.build     # Build Crossplane package with embedded runtime and MRDs
make clean          # Clean build artifacts
```

### Development Workflow
```bash
# 1. Set up provider structure and generate APIs
make generate

# 2. Implement controllers and clients
# Edit files in internal/controller/ and internal/clients/

# 3. Validate changes
make reviewable

# 4. Build and test locally
make build
make test

# 5. Publish to registry
make publish VERSION=v0.1.0 PLATFORMS=linux_amd64
```

## Critical Package Structure

**MANDATORY Directory Layout (v2 Architecture):**
```
provider-namecheap/
├── apis/                           # API definitions and CRD types
│   └── v1beta1/                   # v2 namespaced API version
│       ├── domain_types.go        # Domain resource types
│       ├── dnsrecord_types.go     # DNS record types
│       └── providerconfig_types.go # Provider configuration
├── cmd/provider/
│   └── main.go                    # Main entry point (environment variables)
├── config/                        # Enhanced provider configuration
│   ├── provider/                  # Provider metadata
│   ├── crd/                       # Generated CRDs (namespaced)
│   └── mrd/                       # Managed Resource Definitions
├── internal/
│   ├── controller/                # Controllers for namespaced resources
│   │   ├── domain/               # Domain controller
│   │   └── dnsrecord/            # DNS record controller
│   └── clients/                  # Namecheap API client wrapper
│       └── namecheap/            # Namecheap-specific client
├── package/
│   ├── crossplane.yaml           # REQUIRED - NOT package.yaml!
│   ├── crds/                     # Generated CRDs for packaging
│   └── mrd/                      # Managed Resource Definitions
├── examples/                     # v2 namespaced usage examples
├── cluster/images/provider-namecheap/
│   └── Dockerfile                # MUST use ENTRYPOINT, not CMD
├── build/                        # rossigee/build submodule
└── Makefile                      # Build orchestration
```

## Namecheap API Integration

### Client Implementation
- **Location**: `internal/clients/namecheap/`
- **Authentication**: API key and username authentication
- **Rate Limiting**: Implement proper rate limiting for Namecheap API
- **Error Handling**: Robust error handling with meaningful messages
- **Testing**: Mock clients for unit testing

### API Endpoints
- **Domains API**: Domain registration, renewal, transfer
- **DNS API**: DNS record management (A, AAAA, CNAME, MX, TXT, SRV)
- **SSL API**: SSL certificate management (if implementing)

### Controller Patterns
Each resource type should have:
- Dedicated controller in `internal/controller/`
- External resource reconciliation
- Proper status reporting
- Deletion policy handling
- Cross-reference resolution

## Registry Configuration

**Primary Registry**: `ghcr.io/rossigee/provider-namecheap`
- All published versions use this registry pattern
- Version tags follow semantic versioning (v0.1.0, v0.2.0, etc.)
- Latest tag points to most recent stable release

## Critical Build System Requirements

### 1. Package Metadata File
- **MUST** be named `package/crossplane.yaml` (not `package.yaml`)
- Required for Docker image embedding in .xpkg package
- Missing this file causes "no command specified" container errors

### 2. Dockerfile Requirements
```dockerfile
FROM gcr.io/distroless/static:nonroot
COPY --chmod=0755 bin/linux_amd64/provider /usr/local/bin/provider
ENTRYPOINT ["/usr/local/bin/provider"]
# NEVER use CMD - use ENTRYPOINT
```

### 3. Environment Variable Configuration
In `cmd/provider/main.go`:
```go
// Correct - use environment variables
CertDir: os.Getenv("WEBHOOK_TLS_CERT_DIR")
TLSServerCertDir: os.Getenv("TLS_SERVER_CERTS_DIR")

// Wrong - never hardcode paths
// CertDir: "/tmp/k8s-webhook-server/serving-certs"
```

## API Design Guidelines (v2 Architecture)

### Resource Naming Convention
- **API Group**: `namecheap.m.crossplane.io` (note the `.m.` for namespaced)
- **Version**: `v1beta1` (v2 namespaced resources often reset versioning)
- **Kinds**: Domain, DNSRecord, ProviderConfig
- **Scope**: Namespaced resources for multi-tenant isolation

### Example Resource Definitions (v2 Namespaced)
```yaml
# Domain resource (v2 namespaced)
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: example-com
  namespace: production  # Namespace isolation
spec:
  forProvider:
    domainName: example.com
    registrationYears: 1
    nameservers:
      - ns1.namecheap.com
      - ns2.namecheap.com
  providerConfigRef:
    name: default
  deletionPolicy: Delete

# DNS Record resource (v2 namespaced)
apiVersion: namecheap.m.crossplane.io/v1beta1
kind: DNSRecord
metadata:
  name: www-record
  namespace: production  # Namespace isolation
spec:
  forProvider:
    domain: example.com
    type: A
    name: www
    value: 192.168.1.1
    ttl: 300
  providerConfigRef:
    name: default
  deletionPolicy: Delete

# Managed Resource Definition (MRD) example
apiVersion: meta.pkg.crossplane.io/v1alpha1
kind: ManagedResource
metadata:
  name: domain.namecheap.m.crossplane.io
spec:
  resources:
  - namecheap.m.crossplane.io/v1beta1/Domain
  activation:
    policy: Automatic
```

## Testing Strategy

### Unit Tests
- Mock Namecheap API clients
- Test controller reconciliation logic
- Validate CRD generation and schema
- Error handling scenarios

### Integration Tests
- Test against Namecheap sandbox API (if available)
- End-to-end resource lifecycle testing
- Provider configuration validation

### Commands
```bash
make test                    # Run all unit tests
go test ./internal/...      # Run specific package tests
go test -v -run TestDomain  # Run specific test function
```

## Deployment and Verification

### Build Verification
```bash
# Check Docker image has correct ENTRYPOINT
docker inspect ghcr.io/rossigee/provider-namecheap-amd64 | jq -r '.[0].Config.Entrypoint'
# Should show: ["/usr/local/bin/provider"]

# Verify .xpkg package contains Docker image
tar -tf _output/xpkg/linux_amd64/provider-namecheap-*.xpkg | grep manifest.json
```

### Deployment Commands
```bash
# Apply provider to cluster
kubectl apply -f examples/provider/

# Check provider status
kubectl get providers
kubectl describe provider provider-namecheap

# Verify provider pod is running
kubectl get pods -n crossplane-system | grep namecheap

# Check provider logs
kubectl logs -n crossplane-system deployment/provider-namecheap
```

## Common Issues and Solutions

### Build System Issues
- **"No rule to make target 'lint'"**: Wrong build submodule - ensure using `github.com/rossigee/build`
- **"go mod tidy needed"**: Run `go mod tidy` to update dependencies
- **Go version errors**: Set `GO_REQUIRED_VERSION ?= 1.24` in Makefile

### Runtime Issues
- **"no command specified"**: Missing `package/crossplane.yaml` or wrong Dockerfile ENTRYPOINT
- **TLS certificate errors**: Use environment variables in main.go, not hardcoded paths
- **API authentication failures**: Verify ProviderConfig credentials and Namecheap API key

### Debugging Commands
```bash
# Extract and inspect .xpkg package
tar -tf provider-namecheap.xpkg | head -20

# Check provider configuration
kubectl describe providerconfig default

# Verify resource creation
kubectl get domains,dnsrecords -A
kubectl describe domain example-com
```

## Development Standards

### Code Organization
- Controllers in `internal/controller/` with dedicated subdirectories
- API clients in `internal/clients/namecheap/`
- Types in `apis/v1alpha1/` with proper Go struct tags
- Examples in `examples/` directory with working manifests

### Error Handling
- Always wrap external API errors with context
- Use Crossplane conditions for status reporting
- Implement proper logging with structured fields
- Handle rate limiting and retries gracefully

### Documentation
- Inline code comments for complex logic
- README with setup and usage instructions
- Examples directory with real-world use cases
- API reference documentation generated from Go types

## Version Compatibility

**Crossplane Version**: v1.20+ required (v2 features)
**Go Version**: 1.24+
**Kubernetes**: 1.28+

This provider implements Crossplane v2-only architecture with namespaced resources and modern patterns. No legacy v1 support is provided to reduce complexity and maintenance overhead.