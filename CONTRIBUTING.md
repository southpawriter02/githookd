# Contributing to githookd

Thank you for your interest in contributing! This document covers the development workflow and CI/CD process.

## Development Setup

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [golangci-lint](https://golangci-lint.run/welcome/install-local/) (for linting)
- [GoReleaser](https://goreleaser.com/install/) (optional, for release snapshots)

### Getting Started

```bash
git clone https://github.com/YOUR_ORG/githookd.git
cd githookd

# Build
make build

# Run tests
make test

# Run linter
make lint
```

### Makefile Targets

| Target            | Description                           |
| ----------------- | ------------------------------------- |
| `make build`      | Build binary with version info        |
| `make test`       | Run all tests with race detection     |
| `make test-cover` | Run tests and generate coverage HTML  |
| `make lint`       | Run golangci-lint                     |
| `make clean`      | Remove build artifacts                |
| `make install`    | Install to `$GOPATH/bin`              |
| `make snapshot`   | Build a snapshot release (no publish) |

---

## CI Pipeline

Every push to `main` and every pull request triggers the CI pipeline (`.github/workflows/ci.yml`), which runs:

1. **Lint** — `golangci-lint` on Ubuntu
2. **Test** — `go test -race` across a matrix of:
   - OS: `ubuntu-latest`, `macos-latest`, `windows-latest`
   - Go: `1.23`, `1.24`
3. **Build** — Cross-platform compilation on all three OS runners

All three jobs must pass for the `CI ✓` status check to succeed.

### Additional Workflows

| Workflow                | Trigger             | Purpose                                |
| ----------------------- | ------------------- | -------------------------------------- |
| `ci.yml`                | Push / PR to `main` | Lint, test, build                      |
| `release.yml`           | Tag push (`v*`)     | GoReleaser — binaries + GitHub release |
| `codeql.yml`            | Push / PR + weekly  | Security analysis                      |
| `dependency-review.yml` | PRs to `main`       | Dependency vulnerability check         |

---

## Release Process

Releases are fully automated. To cut a new release:

```bash
# 1. Ensure you're on main with a clean working tree
git checkout main
git pull

# 2. Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"

# 3. Push the tag
git push origin v1.0.0
```

This triggers the release workflow, which:

1. **Runs the full CI pipeline** (lint + test + build)
2. **Builds cross-platform binaries** via GoReleaser:
   - `linux/amd64`, `linux/arm64`
   - `darwin/amd64`, `darwin/arm64`
   - `windows/amd64`, `windows/arm64`
3. **Creates a GitHub Release** with:
   - Categorized changelog (features, fixes, docs, maintenance)
   - Archive files (`.tar.gz` for unix, `.zip` for Windows)
   - SHA-256 checksums
4. **Publishes a Homebrew formula** to the tap repository

### Snapshot Builds

To test the release process locally without publishing:

```bash
make snapshot
# or
goreleaser release --snapshot --clean
```

This creates binaries in `dist/` without pushing anything.

### Release Secrets

The release workflow requires these repository secrets:

| Secret               | Purpose                                          |
| -------------------- | ------------------------------------------------ |
| `GITHUB_TOKEN`       | Auto-provided; creates GitHub releases           |
| `HOMEBREW_TAP_TOKEN` | PAT with `repo` scope on the `homebrew-tap` repo |

---

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for automatic changelog generation:

```
feat: add parallel execution support
fix: resolve symlink resolution on Windows
docs: update CI/CD integration guide
chore: bump Go version to 1.24
ci: add CodeQL analysis workflow
refactor: extract hook resolution logic
test: add coverage for timeout edge cases
```

The GoReleaser changelog groups commits by prefix into Features, Bug Fixes, Documentation, and Maintenance sections.

---

## Branch Protection

We recommend configuring these branch protection rules for `main`:

- ✅ Require status check: **CI ✓**
- ✅ Require pull request reviews
- ✅ Require branches to be up to date before merging
- ✅ Require signed commits (optional)
