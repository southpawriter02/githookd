# 2. Hook Chaining

This document describes the hook chaining feature of the Git Hook Manager.

## 2.1. Intended Functionality

Hook chaining will allow users to define a sequence of hooks that should be executed in a specific order. This is useful for creating complex workflows where the output of one hook is used as the input for another.

Users will be able to define chains in the `.githooksrc.yml` file:

```yaml
# .githooksrc.yml

hooks:
  pre-commit:
    - chain: "lint-and-format"

chains:
  lint-and-format:
    - run: "npm run lint"
    - run: "npm run format"
```

In this example, the `lint-and-format` chain will run the `lint` command first, and if it succeeds, it will run the `format` command.

## 2.2. Requirements

- The tool must be able to parse and execute hook chains defined in the configuration file.
- The tool must ensure that the hooks in a chain are executed in the specified order.
- If a hook in the chain fails, the execution of the chain must be aborted.

## 2.3. Limitations

- Initially, the tool will not support passing data between hooks in a chain directly. Users will have to use temporary files or other mechanisms to share data.
- The hook chaining feature will be limited to hooks within the same Git hook event (e.g., `pre-commit`). Chaining hooks across different events (e.g., from `pre-commit` to `commit-msg`) will not be supported.

## 2.4. Dependencies

- This feature will not introduce any new external dependencies. It will be implemented using the existing core functionality of the Git Hook Manager.
