# 2. Interactive Mode

This document describes the interactive mode for the Git Hook Manager.

## 2.1. Intended Functionality

To improve the user experience, especially for new users, we will provide an interactive mode for the `ghm` CLI. In this mode, the tool will guide the user through the process of setting up and configuring the Git Hook Manager.

The interactive mode will be triggered when the user runs `ghm install` in a project for the first time. It will ask the user a series of questions to help them create their initial configuration file.

## 2.2. Interactive Setup

The interactive setup process will include the following steps:

1. **Detecting the Project Type:** The tool will try to detect the type of project (e.g., Node.js, Python, Go) and suggest a set of common hooks for that project type.
2. **Selecting Hooks:** The user will be able to select which hooks they want to enable from a list of predefined hooks.
3. **Configuring Hooks:** For each selected hook, the user will be able to configure the command to run and other options.
4. **Generating the Configuration File:** Based on the user's answers, the tool will generate the `.githooksrc.yml` file.

## 2.3. Requirements

- The interactive mode must be user-friendly and easy to understand.
- The tool should provide sensible defaults for the configuration options.
- The user should be able to skip the interactive mode and create the configuration file manually.

## 2.4. Limitations

- The initial version of the interactive mode will only be available for the `ghm install` command. We may add interactive modes for other commands in the future.
- The project type detection will be based on the presence of certain files (e.g., `package.json` for Node.js, `requirements.txt` for Python). It may not be accurate for all projects.

## 2.5. Dependencies

- The interactive mode will require a library for creating interactive command-line interfaces. The choice of library will depend on the implementation language (e.g., `survey` for Go, `inquirer` for Python).
