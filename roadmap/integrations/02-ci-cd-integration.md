# 2. CI/CD Integration

This document describes the CI/CD integration for the Git Hook Manager.

## 2.1. Intended Functionality

The Git Hook Manager can be used in a CI/CD pipeline to enforce the same checks that are run locally. This ensures that all code that is merged into the main branch has passed the required quality checks.

We will provide a command to run all the hooks for a specific Git hook event:

```bash
ghm run <hook-name>
```

This command can be used in a CI/CD script to execute the hooks.

## 2.2. Example Usage

Here is an example of how to use the Git Hook Manager in a GitHub Actions workflow:

```yaml
# .github/workflows/ci.yml

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install Git Hook Manager
        run: |
          # Installation command for GHM
      - name: Run pre-commit hooks
        run: ghm run pre-commit
```

## 2.3. Requirements

- The `ghm run` command must be able to execute hooks in a non-interactive environment.
- The tool should provide a clear and concise output that can be easily parsed by CI/CD systems.
- The tool should exit with a non-zero status code if any of the hooks fail.

## 2.4. Limitations

- The initial version of the CI/CD integration will only support running hooks. We will not provide any specific integrations with CI/CD platforms.
- It will be the user's responsibility to install the Git Hook Manager in their CI/CD environment.

## 2.5. Dependencies

- This feature will not introduce any new external dependencies. It will be implemented using the existing core functionality of the Git Hook Manager.
