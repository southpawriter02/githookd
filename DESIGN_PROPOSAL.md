# githookd GUI/UX Enhancement Proposal

This document outlines the proposal for modernizing the User Experience (UX) and User Interface (UI) of `githookd`. The goal is to transform the tool from a basic CLI utility into a rich, interactive developer tool that inspires confidence and ease of use.

## 1. Vision & Goals

*   **Visual Clarity**: Replace plain text with structured, colored, and styled output to make status and errors immediately parseable.
*   **Interactivity**: Reduce the need to manually edit YAML files or remember flags by providing interactive wizards and menus.
*   **Safety**: Introduce "Dry Run" capabilities and clear confirmations to prevent accidental overwrites or misconfigurations.
*   **Modern Developer Experience**: Align with the standards of modern CLI tools (like `gh`, `lazygit`, etc.) using a TUI (Text User Interface).

## 2. Technical Stack Selection

We will utilize the **Charm** ecosystem for Go, as it is the industry standard for modern, beautiful CLIs.

*   **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture for Go).
*   **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss) (CSS-like styling for terminals).
*   **Interactive Forms**: [Huh](https://github.com/charmbracelet/huh) (For wizards and prompts).

## 3. Implementation Specifications

### Phase 1: CLI Modernization

This phase focuses on improving the existing command-line experience without launching a full-screen application.

#### A. Styling & Output (`internal/style`)
*   Create a centralized style package.
*   **Success Messages**: Green checkmarks, bold text.
*   **Error Messages**: Red boxes, clear "Actionable" hints.
*   **Info/Logs**: Dimmed text for low-priority info.

#### B. New Command: `ghm status`
A dashboard-like view printed to the console.

*   **Logic**: Iterates through supported git hooks. Checks if:
    1.  The hook is present in `.githooksrc.yml`.
    2.  The symlink in `.git/hooks/` exists and points to `githookd`.
*   **Output Preview**:
    ```text
    Githookd Status

    ● pre-commit       Active (2 commands)
    ○ pre-push         Configured but not installed
    - commit-msg       Not configured
    ! post-merge       Conflict (symlink points elsewhere)
    ```

#### C. Interactive `install` Wizard
Replace the immediate execution of `ghm install` with an interactive form (using `huh`).

*   **Prompt 1**: "Which hooks would you like to install?" (Multi-select list, pre-selecting those found in config).
*   **Prompt 2**: "Back up existing hooks?" (Yes/No).
*   **Prompt 3**: "Create default config?" (If missing).

#### D. Dry Run & Preview
Add a `--dry-run` flag to `run` and `install` commands.
*   **`ghm run <hook> --dry-run`**: Prints the commands that *would* be executed, showing environment variables and arguments, without actually running them.

### Phase 2: TUI (Text User Interface)

This phase introduces a full-screen interactive mode via a new command: `ghm ui`.

#### A. Layout
*   **Left Column (Sidebar)**: List of Git hooks (up/down navigation). Icons indicating status (Active, Inactive, Error).
*   **Right Column (Content)**:
    *   **Header**: Hook Name & Description.
    *   **Body**: List of commands configured for this hook.
    *   **Footer**: Help keys (`q`: Quit, `e`: Edit, `space`: Toggle).

#### B. Features
1.  **Browse**: Quickly cycle through hooks to see what is configured.
2.  **Toggle**: Press `Space` on a hook in the sidebar to Install/Uninstall the symlink instantly.
3.  **Edit Config**: Press `e` to open a modal form to add/remove commands for the selected hook.
    *   *Implementation*: This will modify the `.githooksrc.yml` file in memory and save on confirm.
4.  **Test Run**: Press `r` to manually trigger the hook (useful for testing `pre-commit` scripts without committing).

## 4. Proposed Roadmap

1.  **Foundation**: Add `lipgloss` dependency and refactor `cmd/ghm` to use a `logger/printer` interface that supports styling.
2.  **Status Command**: Implement `ghm status`.
3.  **Interactive Install**: Refactor `install.go` to use `huh`.
4.  **TUI Core**: Initialize `bubbletea` model for `ghm ui`.
5.  **TUI Features**: Implement file editing and process execution within the TUI.

## 5. Mockup (TUI)

```
+------------------+------------------------------------------------+
|  GITHOOKD        |  PRE-COMMIT                                    |
+------------------+------------------------------------------------+
| ● applypatch-msg |                                                |
| ● pre-applypatch |  Status: ● Active                              |
| > pre-commit     |                                                |
| ○ post-commit    |  Commands:                                     |
| ○ pre-push       |  1. Lint                                       |
|                  |     Run: golangci-lint run                     |
|                  |                                                |
|                  |  2. Test                                       |
|                  |     Run: go test ./...                         |
|                  |                                                |
+------------------+------------------------------------------------+
| ↑/↓: Navigate  Space: Toggle Install  e: Edit  q: Quit            |
+------------------+------------------------------------------------+
```
