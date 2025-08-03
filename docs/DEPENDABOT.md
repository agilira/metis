# Dependabot Configuration

This document describes the Dependabot configuration for Metis.

## Overview

Dependabot automatically creates pull requests to update dependencies in your repository. It can help you keep your dependencies up to date with the latest security patches and features.

## Configuration

The Dependabot configuration is located in `.github/dependabot.yml` and monitors:

### Go Dependencies (`gomod`)
- **Schedule**: Weekly on Monday at 09:00
- **Updates**: Patch and minor versions only
- **Major versions**: Ignored to prevent breaking changes
- **Labels**: `dependencies`, `go`
- **Commit prefix**: `deps`

### GitHub Actions
- **Schedule**: Weekly on Monday at 09:00
- **Updates**: Patch and minor versions only
- **Major versions**: Ignored for critical actions
- **Labels**: `dependencies`, `github-actions`
- **Commit prefix**: `ci`

## Workflow Integration

The `.github/workflows/dependabot.yml` workflow automatically tests Dependabot PRs:

1. **Dependency Verification**: Ensures `go.mod` and `go.sum` consistency
2. **Test Execution**: Runs all tests to verify compatibility
3. **Security Scan**: Performs quick security scan with gosec
4. **PR Comments**: Automatically comments on PRs with test results

## Security Considerations

- **Major version updates**: Manually reviewed to prevent breaking changes
- **Security scans**: All dependency updates are scanned for security issues
- **Test coverage**: All updates must pass existing tests

## Manual Updates

For manual dependency updates:

```bash
# Update all dependencies
go get -u ./...

# Update specific dependency
go get -u github.com/example/package

# Update to specific version
go get github.com/example/package@v1.2.3

# Tidy up go.mod
go mod tidy
```

## Best Practices

1. **Review PRs**: Always review Dependabot PRs before merging
2. **Test locally**: Test major version updates locally before merging
3. **Monitor security**: Pay attention to security-related updates
4. **Update regularly**: Don't let dependencies get too outdated

## Troubleshooting

### Common Issues

1. **Build failures**: Check if new dependency versions are compatible
2. **Security warnings**: Review gosec output for new security issues
3. **Breaking changes**: Major version updates may require code changes

### Disabling Dependabot

To temporarily disable Dependabot:

1. Comment out sections in `.github/dependabot.yml`
2. Or add `ignore` rules for specific dependencies

### Re-enabling Dependabot

1. Uncomment sections in `.github/dependabot.yml`
2. Remove `ignore` rules as needed
3. Dependabot will resume on the next scheduled run 

---

Metis • an AGILira fragment