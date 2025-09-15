# Contributing to provider-namecheap

Thank you for your interest in contributing to provider-namecheap! This document outlines the process for contributing to this Crossplane provider.

## Getting Started

### Prerequisites

- Go 1.24+
- Docker
- Make
- Git with submodules support
- `kubectl` for testing
- `pre-commit` (recommended)

### Setting up Development Environment

1. **Fork and Clone**
   ```bash
   git clone --recursive https://github.com/YOUR_USERNAME/provider-namecheap.git
   cd provider-namecheap
   ```

2. **Install Pre-commit Hooks** (recommended)
   ```bash
   pre-commit install
   ```

3. **Verify Setup**
   ```bash
   make reviewable
   ```

## Development Workflow

### Making Changes

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Follow existing code patterns and conventions
   - Add tests for new functionality
   - Update documentation as needed

3. **Validate Changes**
   ```bash
   # Run all quality checks
   make reviewable

   # Or run specific checks
   make lint
   make test
   make generate
   ```

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add support for SRV records"
   ```

### Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

Examples:
- `feat: add support for TXT record management`
- `fix: resolve DNS record update race condition`
- `docs: update installation instructions`

### Pull Request Process

1. **Update Documentation**
   - Update README.md if adding new features
   - Add/update examples if needed
   - Update API documentation

2. **Ensure Quality**
   - All tests pass (`make test`)
   - No linting errors (`make lint`)
   - Code generation is up to date (`make generate`)

3. **Create Pull Request**
   - Use a descriptive title
   - Reference any related issues
   - Describe what changes were made and why

4. **Address Feedback**
   - Respond to review comments
   - Make requested changes
   - Keep the PR updated with master

## Code Organization

### Directory Structure

```
provider-namecheap/
├── apis/v1beta1/           # API type definitions
├── cmd/provider/           # Main provider entry point
├── internal/
│   ├── clients/namecheap/ # Namecheap API client
│   ├── controller/        # Resource controllers
│   └── version/           # Version information
├── examples/              # Usage examples
├── package/               # Crossplane package metadata
└── cluster/images/        # Docker build configuration
```

### Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Use the existing error handling patterns

### API Design

- Follow Crossplane v2 patterns with namespaced resources
- Use the `.m.` API group format: `namecheap.m.crossplane.io`
- Implement all required resource.Managed interface methods
- Follow Kubernetes API conventions

## Testing

### Unit Tests

- Write unit tests for all new functionality
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for good test coverage

```bash
make test
```

### Integration Testing

For testing with real Namecheap API:

1. **Setup Test Credentials**
   ```bash
   # Create a test secret
   kubectl create secret generic namecheap-test-creds -n crossplane-system \
     --from-literal=credentials='{"apiUser":"...","apiKey":"...","username":"...","clientIP":"..."}'
   ```

2. **Enable Sandbox Mode**
   ```yaml
   apiVersion: namecheap.m.crossplane.io/v1beta1
   kind: ProviderConfig
   metadata:
     name: test
   spec:
     sandboxMode: true  # Important for testing!
   ```

3. **Test Resources**
   ```bash
   kubectl apply -f examples/
   ```

## Debugging

### Enable Debug Logging

```bash
kubectl patch deployment provider-namecheap -n crossplane-system \
  --patch '{"spec":{"template":{"spec":{"containers":[{"name":"package-runtime","args":["--debug"]}]}}}}'
```

### Common Issues

1. **Build Issues**
   - Ensure git submodules are initialized: `git submodule update --init --recursive`
   - Clean build cache: `make clean`

2. **API Issues**
   - Verify Namecheap API credentials
   - Check IP address whitelist in Namecheap account
   - Use sandbox mode for testing

3. **CRD Issues**
   - Regenerate CRDs: `make generate`
   - Check CRD installation: `kubectl get crds | grep namecheap`

## Documentation

### Code Documentation

- Add Go doc comments for public functions and types
- Use examples in documentation where helpful
- Keep comments up to date with code changes

### User Documentation

- Update README.md for new features
- Add examples for new resource types
- Update troubleshooting section as needed

## Release Process

Releases are handled automatically via GitHub Actions:

1. **Create a Tag**
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

2. **Automated Process**
   - CI builds and tests the release
   - Docker images are built and pushed
   - Crossplane packages are published
   - GitHub release is created

## Getting Help

- **Issues**: Check existing [issues](https://github.com/rossigee/provider-namecheap/issues)
- **Discussions**: Use [GitHub Discussions](https://github.com/rossigee/provider-namecheap/discussions)
- **Slack**: Join the Crossplane community slack

## Code of Conduct

This project follows the [Crossplane Code of Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md).

## License

By contributing to provider-namecheap, you agree that your contributions will be licensed under the Apache License 2.0.