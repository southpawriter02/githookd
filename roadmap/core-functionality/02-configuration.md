# 2. Configuration

This document details the configuration mechanism for the Git Hook Manager.

## 2.1. Intended Functionality

The Git Hook Manager will be configured through a file in the root of the project's repository. This configuration file will allow users to define and customize the behavior of the tool. Key configuration options will include:

- **Hook Definitions:** Specifying which hooks to manage and the commands they should execute.
- **Global Settings:** Settings that apply to all hooks, such as logging levels or execution timeouts.
- **Hook-Specific Settings:** Overriding global settings for individual hooks.

## 2.2. Configuration File Format

The configuration file will be named `.githooksrc.yml` and will use the YAML format. We chose YAML for its readability and ability to represent complex data structures.

Here is an example of what the configuration file might look like:

```yaml
# .githooksrc.yml

# Global settings
timeout: 10s
log_level: info

hooks:
  pre-commit:
    - run: "npm run lint"
      description: "Run ESLint"
    - run: "npm run test"
      description: "Run unit tests"
  commit-msg:
    - run: "python scripts/validate_commit_msg.py $1"
      description: "Validate commit message format"
```

## 2.3. Requirements

- The tool must parse the `.githooksrc.yml` file to configure the hooks.
- The tool should provide clear error messages if the configuration file is malformed or contains invalid options.
- The configuration file should be version-controlled, allowing teams to share the same hook configuration.

## 2.4. Limitations

- Initially, the tool will only support a single configuration file at the root of the repository. We may consider per-user or per-directory configuration files in the future.
- The configuration options will be limited to what is described in this document. More advanced configuration options will be added as new features are developed.

## 2.5. Dependencies

- The tool will need a YAML parser library to read the configuration file. The choice of library will depend on the implementation language.
