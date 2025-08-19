# 1. Parallel Execution

This document describes the parallel execution feature of the Git Hook Manager.

## 1.1. Intended Functionality

To speed up the execution of hooks, especially in projects with many checks, the Git Hook Manager will support running hooks in parallel. Users will be able to configure which hooks can be run in parallel and the maximum number of parallel processes.

This will be controlled by a `parallel` key in the `.githooksrc.yml` file:

```yaml
# .githooksrc.yml

hooks:
  pre-commit:
    - run: "npm run lint"
      parallel: true
    - run: "npm run test"
      parallel: true
```

In this example, the `lint` and `test` commands will be run in parallel.

## 1.2. Requirements

- The tool must be able to execute multiple hook scripts concurrently.
- The tool should manage the output of the parallel processes to avoid interleaved and unreadable logs.
- The tool must correctly handle the exit codes of the parallel processes. If any of the parallel hooks fail, the entire hook chain should fail.
- The user should be able to configure the maximum number of parallel processes.

## 1.3. Limitations

- Not all hooks can be run in parallel. For example, hooks that modify the same files can lead to race conditions. It will be the user's responsibility to determine which hooks are safe to run in parallel.
- The parallel execution feature will be disabled by default. Users will need to explicitly enable it in the configuration file.

## 1.4. Dependencies

- The implementation of this feature will likely require a library for managing concurrent processes. The choice of library will depend on the implementation language.
