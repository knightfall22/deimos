# Project Context

This is a Terminal User Interface (TUI) application built in Go. It serves as a chat interface for various AI models (Ollama, OpenAI, Claude, Gemini). The application features a split-pane layout with a model selection list and an input textarea.

## Tech Stack

- Language: Go (Golang)
- TUI Framework: `charm.land/bubbletea/v2`
- Components: `charm.land/bubbles/v2` (List, Textarea)
- Styling: `charm.land/lipgloss/v2`

## Architecture & State Management

- **Pattern:** Strict adherence to the Elm architecture (Model, Update, View).
- **State (`Model`):** All application state must be immutable within the `Update` loop.
- **Side Effects:** Never perform blocking operations (like network requests to AI APIs) directly inside the `Update` function. All I/O, timers, or long-running tasks MUST be wrapped in a `tea.Cmd` and return a `tea.Msg` upon completion.
- **Routing:** Message routing should check the `activePane` state to ensure keystrokes (like typing in the textarea vs. scrolling the list) are sent to the correct active `bubbles` component.

## Coding Standards (Built via cc-golang-skills)

- **Idiomatic Go:** Write clean, senior-level Go. Utilize early returns to avoid deep nesting.
- **Error Handling:** Errors should be treated as values and explicitly handled. For Bubble Tea, map underlying errors to a specific `errMsg` type to be displayed in the UI.
- **Styling:** Keep all UI styling logic (`lipgloss.NewStyle()`) isolated in the `styles` struct or initialization functions. Do not inline styles within the `View()` function to prevent unnecessary memory allocations on every render frame.
- **Concurrency:** Ensure any goroutines spawned via `tea.Cmd` are safely managed and do not introduce race conditions on the main model.

## Commands

- Run application: `go run main.go`
- Format code: `go fmt ./...`
- Update dependencies: `go get -u ./... && go mod tidy`

## Boundaries & "Never Do"

- NEVER block the `Update` loop. A frozen UI is a critical failure.
- DO NOT mix `View` logic into the `Update` function. `View` must remain a pure function of the `Model` state.
- DO NOT downgrade the `v2` Charm library imports to `v1`. Ensure all new bubble/lipgloss imports explicitly use `/v2`.
- NEVER hardcode API keys or secrets in the source code; assume they will be injected via environment variables later.
