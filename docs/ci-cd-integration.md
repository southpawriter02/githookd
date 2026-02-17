# CI/CD Integration Guide

githookd can run your configured hooks in any CI/CD environment, ensuring the same quality checks that run locally are enforced on every push and pull request.

## Quick Start

The simplest way to use githookd in CI is with the `ghm run` command:

```bash
ghm run pre-commit
```

This executes all commands defined under `pre-commit` in your `.githooksrc.yml`, exits with a non-zero status if any hook fails, and produces structured output that CI systems can parse.

---

## GitHub Actions

### Using the Setup Action

githookd provides a reusable composite action that downloads and installs the correct `ghm` binary for your runner's OS and architecture:

```yaml
# .github/workflows/hooks.yml
name: Run Hooks

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  hooks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Install githookd
      - uses: githookd/githookd/.github/actions/setup-ghm@main
        with:
          version: latest # or a specific tag like 'v1.0.0'

      # Run your hooks
      - name: Run pre-commit hooks
        run: ghm run pre-commit

      - name: Run pre-push hooks
        run: ghm run pre-push
```

### Action Inputs

| Input          | Default               | Description                                 |
| -------------- | --------------------- | ------------------------------------------- |
| `version`      | `latest`              | Version to install (`latest` or a `v*` tag) |
| `github-token` | `${{ github.token }}` | Token for GitHub API (avoids rate limiting) |

### Action Outputs

| Output    | Description                       |
| --------- | --------------------------------- |
| `version` | The installed version tag         |
| `path`    | Absolute path to the `ghm` binary |

### Manual Installation

If you prefer not to use the composite action, you can install `ghm` directly:

```yaml
jobs:
  hooks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install githookd
        run: |
          VERSION="v1.0.0"
          curl -sSfL "https://github.com/YOUR_ORG/githookd/releases/download/${VERSION}/githookd_${VERSION#v}_linux_amd64.tar.gz" \
            | tar xz -C /usr/local/bin ghm

      - name: Run pre-commit hooks
        run: ghm run pre-commit
```

---

## GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - quality

hooks:
  stage: quality
  image: golang:1.24
  before_script:
    - curl -sSfL "https://github.com/YOUR_ORG/githookd/releases/latest/download/githookd_linux_amd64.tar.gz" \
      | tar xz -C /usr/local/bin ghm
  script:
    - ghm run pre-commit
```

---

## CircleCI

```yaml
# .circleci/config.yml
version: 2.1

jobs:
  hooks:
    docker:
      - image: cimg/go:1.24
    steps:
      - checkout
      - run:
          name: Install githookd
          command: |
            curl -sSfL "https://github.com/YOUR_ORG/githookd/releases/latest/download/githookd_linux_amd64.tar.gz" \
              | tar xz -C /usr/local/bin ghm
      - run:
          name: Run pre-commit hooks
          command: ghm run pre-commit

workflows:
  quality:
    jobs:
      - hooks
```

---

## Azure Pipelines

```yaml
# azure-pipelines.yml
trigger:
  branches:
    include: [main]

pool:
  vmImage: "ubuntu-latest"

steps:
  - checkout: self

  - script: |
      curl -sSfL "https://github.com/YOUR_ORG/githookd/releases/latest/download/githookd_linux_amd64.tar.gz" \
        | tar xz -C /usr/local/bin ghm
    displayName: Install githookd

  - script: ghm run pre-commit
    displayName: Run pre-commit hooks
```

---

## Environment Variables

When githookd executes hooks, it sets the following environment variables that your scripts can use:

| Variable        | Description                          |
| --------------- | ------------------------------------ |
| `GHM_HOOK_NAME` | The name of the hook being executed  |
| `GHM_ROOT`      | The root directory of the repository |

---

## Tips

- **Keep hooks fast.** CI runs every hook sequentially. If your pipeline is slow, consider splitting expensive checks into separate CI jobs rather than running them all as hooks.
- **Use the same config.** The whole point is that `.githooksrc.yml` is version-controlled — what runs locally is exactly what runs in CI.
- **Exit codes matter.** `ghm run` exits with code 1 if any hook fails, which correctly marks the CI step as failed.
- **Timeouts.** The default timeout (30s) may be too short for CI. Set `timeout: 5m` on individual commands in `.githooksrc.yml` for slower tasks, or `timeout: none` to disable entirely.
