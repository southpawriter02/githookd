# 4. Automatic Updates

This document describes the automatic update feature for the Git Hook Manager.

## 4.1. Intended Functionality

To ensure that users are always using the latest version of the Git Hook Manager, we will provide an automatic update feature. The tool will periodically check for new versions and, if one is available, prompt the user to update.

This feature will be enabled by default, but users will be able to disable it in the configuration file.

## 4.2. Update Process

The update process will be as follows:

1. The Git Hook Manager will check for a new version in the background.
2. If a new version is found, the tool will display a message to the user, informing them of the new version and asking them to update.
3. If the user agrees, the tool will download the new version and replace the current executable.

## 4.3. Requirements

- The update process must be secure. The tool must verify the integrity of the downloaded update to prevent man-in-the-middle attacks.
- The update process must be robust. The tool should be able to handle network errors and other issues that may occur during the update process.
- The tool should not update itself without the user's consent.

## 4.4. Limitations

- The automatic update feature will only be available for installations done through our official distribution channels (e.g., Homebrew, npm).
- The tool will not be able to update itself if it was installed through a package manager that handles updates on its own (e.g., `apt`).

## 4.5. Dependencies

- The automatic update feature will require a server to host the new versions of the tool. This will likely be the same server that hosts the plugin registry.
