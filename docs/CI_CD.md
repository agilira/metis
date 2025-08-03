# CI/CD Pipeline Documentation

## Overview

Metis uses GitHub Actions to automate the Continuous Integration and Continuous Deployment process. The pipeline is designed to ensure code quality, security, and automated distribution.

## Workflows

### 1. CI/CD Pipeline (`.github/workflows/ci-cd.yml`)

**Trigger:** Push to `main`/`develop` and Pull Request to `main`

**Jobs:**
- **Test & Lint**: Runs tests, linting, and static analysis
- **Security Scan**: Security scanning with gosec
- **Build**: Multi-platform compilation
- **Benchmark**: Performance benchmark execution
- **Release**: Automatic release creation (only on `main`)

### 2. Pull Request Check (`.github/workflows/pr-check.yml`)

**Trigger:** Pull Request to `main`/`develop`

**Jobs:**
- **Quick Check**: Fast checks for PR (test, vet, fmt, staticcheck, security)

### 3. Security Scan (`.github/workflows/security.yml`)

**Trigger:** Push to `main`/`develop` and Pull Request to `main`

**Jobs:**
- **Security Scan**: Detailed security scanning

## Quality Checks

### Tests
- Execution of all unit tests
- Tests with race condition detection
- Code coverage (target: 90%)
- Automatic upload to Codecov

### Linting
- `go vet`: Code correctness checks
- `go fmt`: Formatting verification
- `golint`: Style checks
- `staticcheck`: Advanced static analysis

### Security
- `gosec`: Security vulnerability scanning
- Exclusion of documented false positives:
  - G103: unsafe.Pointer (performance-critical hashing)
  - G115: integer overflow (bounds checking implemented)
  - G404: weak random (performance profiler)
  - G304: file inclusion (path validation implemented)

## Multi-Platform Build

### Generated Binaries
- **Linux**: amd64, arm64
- **Windows**: amd64
- **macOS**: amd64, arm64

### Tools
- `metis-cli`: CLI tool for configuration
- `profiler`: Performance profiling tool

## Automatic Releases

### Conditions
- Push to `main` branch
- All previous jobs completed successfully

### Process
1. Automatic creation of `v{run_number}` tag
2. GitHub release generation
3. Multi-platform binary upload
4. Automatic notifications

## Configuration

### Codecov (`.codecov.yml`)
- Coverage target: 90%
- Threshold: 5%
- Ignores test files and examples

### Gosec (`.gosec`)
- Configuration to exclude false positives
- Documentation of exclusions

## Monitoring

### Badges
- Build Status
- Test Coverage
- Security Scan
- Release Status

### Notifications
- Slack/Discord integration (configurable)
- Email notifications for releases
- GitHub notifications for PRs

## Troubleshooting

### Common Issues

1. **Failed Tests**
   - Check code coverage
   - Verify race conditions
   - Update tests for new features

2. **Linting Errors**
   - Run `go fmt ./...`
   - Fix golint warnings
   - Resolve staticcheck issues

3. **Security Warnings**
   - Verify if they are false positives
   - Implement security checks
   - Update exclusions if necessary

4. **Build Failures**
   - Check dependencies
   - Verify multi-platform compatibility
   - Update Go version if necessary

### Local Commands

```bash
# Run all checks locally
make ci

# Tests only
go test -v ./...

# Linting only
go vet ./...
gofmt -s -l .
golint ./...
staticcheck ./...

# Security only
gosec -exclude=G103,G115,G404,G304 ./...

# Multi-platform build
make build-all
```

## Contributing

1. Create branch from `develop`
2. Implement features with tests
3. Run local checks
4. Create Pull Request
5. Wait for CI/CD approval
6. Merge after review

## Support

For CI/CD issues:
1. Check GitHub Actions logs
2. Verify local configuration
3. Contact maintainers
4. Create GitHub issue 

---

Metis • an AGILira fragment