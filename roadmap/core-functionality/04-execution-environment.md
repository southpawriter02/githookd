# 4. Execution Environment

This document describes the execution environment for the hooks managed by the Git Hook Manager.

## 4.1. Intended Functionality

When a Git hook is triggered, the Git Hook Manager will execute the commands defined in the `.githooksrc.yml` file. The execution environment will have the following characteristics:

- **Working Directory:** The hooks will be executed in the root of the Git repository.
- **Environment Variables:** The tool will expose a set of environment variables to the hook scripts, such as `GHM_HOOK_NAME` (the name of the hook being executed) and `GHM_ROOT` (the root of the repository).
- **Standard Input/Output:** The standard input, output, and error streams of the hook scripts will be managed by the Git Hook Manager. This allows for capturing and logging the output of the hooks.

## 4.2. Hook Execution

The Git Hook Manager will execute the hooks in the order they are defined in the configuration file. If a hook fails (i.e., exits with a non-zero status code), the execution of the remaining hooks in the chain will be aborted, and the Git operation will be stopped.

## 4.3. Requirements

- The execution environment must be consistent and predictable across different platforms.
- The tool must correctly handle the exit codes of the hook scripts.
- The tool should provide a way to pass arguments from the Git hook to the hook scripts. For example, the `commit-msg` hook receives the path to the commit message file as an argument.

## 4.4. Limitations

- Initially, the tool will not support running hooks in a containerized environment (e.g., Docker). This is a feature we may consider in the future.
- The execution environment will not provide any sandboxing or security features. Users should be careful about the hooks they run, especially those from untrusted sources.

## 4.5. Dependencies

- The execution of the hooks will depend on the user's shell environment (e.g., Bash, Zsh). The tool should be designed to be as shell-agnostic as possible.
