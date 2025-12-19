# Discord Bot - Go Implementation (AGENTS.md)

This document defines the operational guidelines, project architecture, and development standards for AI assistants participating in this project.

## AI Assistant Operational Guidelines (Efficiency Optimized)

### 1. Interaction Strategy
- **Single Best Option**: Always provide only the one best technical solution. Do not provide multiple options or alternatives unless specifically asked.
- **Logic Focused**: Only provide the core logic or the specific function requested.
- **No Pre-flight Checks**: I will handle `go build`, linting, and testing locally. Do not waste tokens explaining syntax or verifying builds.
- **Skip Explanations**: Omit all "Here is the code..." or "This function works by..." commentary.

### 2. Technical Context
- **Language**: Golang (Idiomatic Go).
- **Library**: `discordgo` (Assume latest stable version).
- **Style**: Direct and idiomatic Go. Use existing project patterns for error handling.
- **Imports**: Only list new imports required for the snippet; do not list the entire import block unless requested.

### 3. Token Saving Rules
- **Ignore Boilerplate**: Do not include `package main` or standard `func main()` initialization unless those specific parts are being modified.
- **Silence Mode**: Do not apologize for errors or give warnings about Discord API rate limits.


---

## Project Structure

```text
discord-bot/
├── main.go                      # Application entry point
├── resources/                   # Embedded resources (translations, assets)
│   ├── resources.go            # Resource embed declarations (centralized)
│   └── i18n/                   # Internationalization files (*.json)
├── internal/                    # Private application code
│   ├── bot/                    # Bot core logic and lifecycle
│   ├── i18n/                   # Internationalization system (Singleton)
│   ├── config/                 # Configuration management (Singleton)
│   ├── interactions/           # Component & Modal router (Singleton)
│   ├── events/                 # Discord event handlers
│   └── commands/               # Slash command system
│       ├── command.go          # Command interface definition
│       ├── registry.go         # Command registry (Factory pattern)
│       ├── sync.go             # Command synchronization utility
│       └── [module]/           # Individual command modules (ping, help, game, etc.)
├── .air.toml                   # Hot reload configuration
└── go.mod                      # Go module definition
```

---

## Architecture Principles

### 1. Command Independence
- Each command is a self-contained module in its own directory under `internal/commands/`.
- Commands do NOT directly depend on other commands. Shared functionality goes into `internal/config/` or `pkg/`.
- All commands must implement the `Command` interface:
  ```go
  type Command interface {
      // Definition returns the slash command definition for Discord registration
      Definition() *discordgo.ApplicationCommand
      // Execute runs the command logic when invoked
      Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error
  }
  ```

### 2. Interaction Routing
- **Slash Commands (`commands/`)**: Managed by `commands.Registry`. Requires Discord API registration.
- **Components & Modals (`interactions/`)**: Managed by `interactions.Router` (Singleton). Routes by `customID` prefix matching (e.g., `game_bullsandcows_`).
- **Registration**: Interactive modules register handlers via `init()` (triggered by package import).

### 3. Internationalization (i18n)
- **Detection**: Automatically detects user's Discord language on interaction.
- **Supported Locales**: Traditional Chinese (`zh-TW`), English (`en-US`).
- **Usage**:
  ```go
  // Get user's locale
  locale := i18n.GetUserLocaleFromInteraction(i)
  // Translation lookup
  title := i18n.T(locale, "game.bullsandcows.title")
  ```

### 4. Resource Management
- **No Relative Paths**: Never use `../` for file paths in code.
- **Centralized Embedding**: All `embed.FS` declarations must be in the `resources/` package.
- **Path Constants**: Define path constants in the same file as embed declarations.

---

## Development Guide

### Adding New Commands
1. Create a directory in `internal/commands/`.
2. Implement the `Command` interface.
3. Register the command in `internal/bot/bot.go` within `registerCommands()`.

### Adding Interactive Games
1. Create a directory in `internal/commands/game/games/`.
2. Implement game logic and `register.go` (using `init()` to register with the interaction router).
3. Anonymous import the package in `internal/commands/game/game.go` to trigger `init()`.

### Best Practices
- **Dependency Injection (DI)**: Inject minimal interfaces via constructors, not heavy global references.
- **Structured Logging**: Use `log/slog` with key-value pairs for context.
- **Configuration**: Access config via `config.Get()`, never read environment variables directly.
- **Error Handling**: Return errors from `Execute()`; the registry handles user-facing error messages.

---

## Extra Instructions

### 1. Commit Message Guidelines
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:`: New features
- `fix:`: Bug fixes
- `docs:`: Documentation changes
- `refactor:`: Code refactoring (no feature/fix)
- `chore:`: Maintenance tasks, dependency updates

### 2. Security Considerations
- **Input Validation**: Always validate user input from Discord interactions.
- **Permissions**: Check if the user has required permissions before executing sensitive commands.
- **Rate Limiting**: Be mindful of Discord API rate limits when sending multiple messages.

### 3. Deployment
- **Binary**: Build the binary using `go build -o discord-bot`.
- **Sensitive Data**: Never output or ask for Discord Tokens/Secrets. Use placeholder variable names from existing code.
- **Environment**: Ensure `.env` is configured on the host machine (but NEVER commit `.env` or read, output the content in `.env`).
- **Process Management**: Use `systemd` or `pm2` to keep the bot running in production.

---

## Common Commands
- **Development**: `air` (Hot reload) or `go run main.go`
- **Testing**: `go test ./...`
- **Building**: `go build -o discord-bot`
