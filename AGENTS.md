# Discord Bot - Go Implementation

A modern, modular Discord bot written in Go with a clean architecture that separates commands and provides extensibility.

## AI Assistant Operational Guidelines (Efficiency Optimized)

### 1. Interaction Strategy
- **Single Best Option**: Always provide only the one best technical solution. **Do not provide multiple options or alternatives** unless specifically asked.
- **Logic Focused**: Only provide the core logic or the specific function requested.
- **No Pre-flight Checks**: I will handle `go build`, linting, and testing locally. Do not waste tokens explaining syntax or verifying builds.
- **Diffs & Snippets**: Provide only the modified code lines. Never output the entire file unless it's a new file creation.
- **Skip Explanations**: Omit all "Here is the code..." or "This function works by..." commentary or anything like that.

### 2. Technical Context
- **Language**: Golang.
- **Library**: discordgo (Assume latest stable version).
- **Style**: Direct and idiomatic Go. Use existing project patterns for error handling.
- **Imports**: Only list new imports required for the snippet; do not list the entire import block unless requested.

### 3. Token Saving Rules
- **Ignore Boilerplate**: Do not include `package main` or standard `func main()` initialization unless those specific parts are being modified.
- **Silence Mode**: Do not apologize for errors or give warnings about Discord API rate limits—I am aware of the constraints.
- **Strict Context Control**: Only read content from files explicitly referenced with `@` OR the specific **lines currently selected/highlighted** in the editor.
- **No Background Scanning**: Do not scan the directory map or unreferenced files to "gain context" unless I explicitly command a full-repo search.

### 4. Environment
- SENSITIVE: Never output or ask for Discord Tokens/Secrets. Use placeholder variable names from existing code. Never read `.env` file.

## Project Structure

```
discord-bot/
├── main.go                      # Application entry point
├── resources/                   # Embedded resources (translations, assets)
│   ├── resources.go            # Resource embed declarations (centralized)
│   └── i18n/                   # Internationalization files
│       ├── zh-TW.json          # Traditional Chinese translations
│       └── en-US.json          # English translations
├── internal/                    # Private application code
│   ├── bot/                    # Bot core logic
│   │   └── bot.go              # Bot initialization and lifecycle management
│   ├── i18n/                   # Internationalization system
│   │   ├── locale.go           # Locale detection and storage (Singleton)
│   │   └── translator.go       # Translation loading and retrieval
│   ├── commands/               # Slash command system
│   │   ├── command.go          # Command interface definition
│   │   ├── registry.go         # Command registry (factory pattern)
│   │   ├── sync.go             # Command synchronization utility
│   │   ├── ping/               # Ping command module
│   │   │   └── ping.go
│   │   ├── help/               # Help command module
│   │   │   └── help.go
│   │   ├── reload/             # Reload command module
│   │   │   └── reload.go
│   │   └── game/               # Game commands
│   │       ├── game.go         # Game command router
│   │       └── games/
│   │           └── bullsandcows/  # Bulls and Cows game
│   │               ├── handler.go   # Discord interaction handlers
│   │               ├── state.go     # Game state management
│   │               └── register.go  # Interaction registration
│   ├── interactions/           # Component & Modal router
│   │   └── router.go           # Interaction router (Singleton)
│   ├── events/                 # Discord event handlers
│   │   ├── events.go           # Event handler registration
│   │   └── guild.go            # Guild-related events
│   └── config/                 # Configuration management
│       └── config.go           # Environment variable loading (Singleton)
├── .env.example                # Environment variable template
├── .gitignore                  # Git ignore rules
├── .air.toml                   # Hot reload configuration
├── go.mod                      # Go module definition
└── go.sum                      # Dependency checksums
```

## Architecture Overview

```
Bot Initialization Flow:
  1. config.Load() → Singleton configuration
  2. i18n.LoadTranslations() → Load resource files
  3. bot.New()
     ├─ commands.NewRegistry() → Create command registry
     ├─ bot.registerCommands() → Register all commands
     ├─ game import → triggers bullsandcows.init()
     │  └─ interactions.GetRouter() → Register handlers
     └─ events.NewHandler() → Setup Discord event listeners

Design Patterns:
  • Singleton:  config, i18n.LocaleStore, interactions.Router
  • Factory:    commands.Registry, individual commands
  • Strategy:   Command interface for polymorphic execution
  • Observer:   Discord event handlers
  • Router:     Prefix-based interaction routing
```

## Architecture Principles

### 1. Command Independence
Each command is a self-contained module in its own directory under `internal/commands/`:
- Commands do NOT directly depend on other commands
- Shared functionality goes into `internal/config/` or `pkg/`
- Commands implement the `Command` interface defined in `command.go`

### 2. Slash Commands
This bot uses Discord's modern Slash Commands system:
- Commands appear in the `/` menu automatically
- No need to remember command prefixes
- Type-safe with autocomplete support
- Commands are synced with Discord on bot startup

### 3. Command Interface
All commands must implement:
```go
type Command interface {
    // Definition returns the slash command definition for Discord registration
    Definition() *discordgo.ApplicationCommand

    // Execute runs the command logic when invoked
    Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error
}
```

### 4. Interaction Routing
The bot uses two separate systems for handling user interactions:

#### Slash Commands (`commands/`)
- **Pattern**: Factory pattern - created and managed by bot
- **Purpose**: Handle `/command` style interactions
- **Registry**: `commands.Registry` manages command registration
- **Handler**: `registry.HandleInteraction()` routes to command Execute()

#### Components & Modals (`interactions/`)
- **Pattern**: Singleton pattern - global router accessed via `GetRouter()`
- **Purpose**: Handle button clicks, select menus, and modal submissions
- **Router**: `interactions.Router` routes by customID prefix matching
- **Registration**: Games register handlers via `init()` (triggered by import)
- **Example**: `game_bullsandcows_guess_` prefix routes to Bulls & Cows handler

**Why two systems?**
- Slash commands need Discord API registration (sync to guilds)
- Components/modals are ephemeral, only need runtime routing
- Decoupling allows independent testing and development

### 5. Event Handling
Discord events are handled separately in the `internal/events/` package:
- **Modular design** - Each event category has its own file (e.g., `guild.go`)
- **Centralized registration** - All events registered via `events.NewHandler()`
- **Separation of concerns** - Events only handle Discord lifecycle, not business logic
- **Auto-sync**: Guild join events automatically sync commands to new servers

### 6. Configuration Management
- Centralized configuration in `internal/config/`
- Uses singleton pattern for global access
- Commands access config via `config.Get()`
- No direct environment variable access in commands

### 7. Dependency Injection & Decoupling
Commands are designed to be loosely coupled using function injection:

**Help Command** - Uses callback for command list:
```go
help.New(func() []help.CommandInfo {
    // Returns only Name and Description, not full Command objects
    return []help.CommandInfo{{Name: "ping", Description: "..."}}
})
```

**Reload Command** - Uses callback that returns count:
```go
reload.New(func() (int, error) {
    // Returns command count after sync, not Registry reference
    return commandCount, nil
})
```

**Benefits:**
- Commands don't hold references to entire Registry
- Easier to test with mock functions
- Clear, minimal dependencies

### 8. Internationalization (i18n)
The bot supports multiple languages with automatic locale detection:
- **Supported locales**: Traditional Chinese (zh-TW), English (en-US)
- **Automatic detection**: Detects user's Discord language on first interaction
- **Persistent storage**: Remembers user's language preference for future interactions
- **Fallback strategy**: Unsupported locales default to English
- **Centralized translations**: All translation files in `resources/i18n/`

**Architecture:**
```
User Interaction → Locale Detection (i18n.GetUserLocaleFromInteraction)
                → Translation Lookup (i18n.T / i18n.Tf)
                → Localized Response
```

**Usage in commands:**
```go
// Get user's locale
locale := i18n.GetUserLocaleFromInteraction(i)

// Simple translation
title := i18n.T(locale, "game.bullsandcows.title")

// Translation with format arguments
message := i18n.Tf(locale, "game.bullsandcows.attempts", currentAttempts, maxAttempts)
```

**Adding new languages:**
1. Add locale constant in `internal/i18n/locale.go`
2. Create translation file in `resources/i18n/` (e.g., `ja-JP.json`)
3. Add locale to `locales` array in `internal/i18n/translator.go`
4. Update `convertToSupportedLocale()` mapping

## Getting Started

### Prerequisites
- Go 1.21 or later
- A Discord Bot Token ([Get one here](https://discord.com/developers/applications))

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

3. Edit `.env` and add your Discord token:
   ```
   DISCORD_TOKEN=your_token_here
   BOT_PREFIX=!
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Run the bot:
   ```bash
   go run main.go
   ```

## Adding New Commands

### Step 1: Create Command Directory
```bash
mkdir internal/commands/yourcommand
```

### Step 2: Implement Command Interface
Create `internal/commands/yourcommand/yourcommand.go`:

```go
package yourcommand

import "github.com/bwmarrin/discordgo"

type Command struct {
    // Add any dependencies here (e.g., database, API clients)
}

func New() *Command {
    return &Command{}
}

func (c *Command) Definition() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
        Name:        "yourcommand",
        Description: "Description of your command",
        // Add options here if needed (arguments, choices, etc.)
    }
}

func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
    // Your command logic here
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "Response message",
        },
    })
}
```

### Step 3: Register Command
Edit `internal/bot/bot.go` and add your command to `registerCommands()`:

```go
import "hiei-discord-bot/internal/commands/yourcommand"

func (b *Bot) registerCommands() {
    // ... existing commands ...
    b.registry.Register(yourcommand.New())
}
```

### Step 4: Reload Commands (Optional)
After adding a new command, you can use `/reload` in Discord to refresh the command list without restarting the bot.

## Available Commands

### General Commands
- `/ping` - Check bot responsiveness and latency
- `/help` - Display all available commands
- `/reload` - Reload all slash commands (useful during development)

### Game Commands
- `/game bullsandcows [difficulty]` - Play Bulls and Cows number guessing game
  - `easy` - 4 unique digits (e.g., 1234)
  - `hard` - Digits can repeat (e.g., 1123)
  - No arguments - Show game rules and instructions

## Dependencies

- [discordgo](https://github.com/bwmarrin/discordgo) v0.29.0 - Discord API wrapper
- [godotenv](https://github.com/joho/godotenv) v1.5.1 - Environment variable loading
- `log/slog` - Go 1.21+ official structured logging (standard library)

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DISCORD_TOKEN` | Discord bot token | - | Yes |

## Development

### Running in Development

#### Standard Mode
```bash
go run main.go
```

#### Hot Reload Mode (Recommended)
Using `air` for automatic rebuild and restart when code changes:

1. Install air (if not already installed):
   ```bash
   go install github.com/air-verse/air@latest
   ```

2. Run with hot reload:
   ```bash
   air
   ```

The bot will automatically rebuild and restart when you save changes to `.go` files. Configuration is in [.air.toml](.air.toml).

**Benefits:**
- No need to manually stop, rebuild, and restart
- Faster development iteration
- Build errors shown immediately
- Temporary builds stored in `tmp/` directory

### Building
```bash
go build -o discord-bot
```

### Running Tests
```bash
go test ./...
```

## Best Practices

### Command Development
1. **Keep commands independent** - No direct imports between command packages
2. **Use Dependency Injection (DI)** - Inject dependencies via constructor, not global references
   ```go
   // ✅ GOOD - Inject minimal interface
   help.New(func() []help.CommandInfo { ... })

   // ❌ BAD - Hold heavy reference
   help.New(registry *commands.Registry)
   ```
3. **Minimize dependencies** - Only inject what's needed (e.g., data, not entire objects)
4. **Error handling** - Return errors from `Execute()`, registry handles user-facing messages
5. **Configuration** - Access config via `config.Get()`, never read env vars directly
6. **Logging** - Use `log/slog` for structured logging with key-value pairs

### Interaction Handlers (Buttons/Modals)
1. **Register via init()** - Let package import trigger registration
   ```go
   // In bullsandcows/register.go
   func init() {
       router := interactions.GetRouter()
       router.RegisterComponent("game_bullsandcows_", handler)
   }
   ```
2. **Use prefix matching** - CustomID format: `prefix_action_data`
3. **Import in game.go** - Trigger init() by importing the game package
   ```go
   import "hiei-discord-bot/internal/commands/game/games/bullsandcows"
   ```

### Registry Patterns
1. **Singleton for global state** - config, i18n.LocaleStore, interactions.Router
2. **Factory for managed instances** - commands.Registry
3. **Don't mix patterns** - Be consistent within a system

### Code Organization
- `internal/` - Private application code that cannot be imported by other projects
- `main.go` - Single application entry point in project root

### Resource Management
1. **No relative paths in code** - Never use `../` for file paths
2. **Centralized embedded resources** - All `embed.FS` declarations must be in `resources/` package
3. **Path constants** - Define path constants in the same file as embed declarations (e.g., `resources.I18nBasePath`)

**Example:**
```go
// ✅ CORRECT - In resources/resources.go
package resources

import "embed"

//go:embed i18n/*.json
var I18n embed.FS

const I18nBasePath = "i18n"

// ✅ CORRECT - Usage in other packages
import "hiei-discord-bot/resources"

func LoadData() error {
    data, err := resources.I18n.ReadFile(resources.I18nBasePath + "/file.json")
    // ...
}

// ❌ WRONG - Relative paths
//go:embed ../../resources/i18n/*.json
var I18n embed.FS

// ❌ WRONG - Hardcoded paths in usage
data, err := fs.ReadFile("../../resources/file.json")
```

### Logging
The project uses Go's official `log/slog` package for structured logging:
```go
import "log/slog"

// Info logging with structured fields
slog.Info("Registered command", "name", "ping")

// Error logging with context
slog.Error("Failed to execute", "command", cmdName, "error", err)
```

### Adding Discord Event Handlers
Discord events (guild joins, messages, reactions, etc.) are handled in `internal/events/`:

1. **Create event handler file** - Add a new file for the event category (e.g., `message.go`)
2. **Implement handler method** - Add method to `Handler` struct:
   ```go
   // In internal/events/message.go
   func (h *Handler) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
       slog.Info("Message received", "content", m.Content)
   }
   ```
3. **Register handler** - Add to `Register()` in `events.go`:
   ```go
   func (h *Handler) Register() {
       h.session.AddHandler(h.OnGuildCreate)
       h.session.AddHandler(h.OnMessageCreate) // Add new handler
   }
   ```

### Adding Interactive Games
To add a new game with buttons/modals:

1. **Create game directory**: `internal/commands/game/games/yourgame/`
2. **Implement game logic**: `handler.go`, `state.go`
3. **Create registration file**: `register.go`
   ```go
   package yourgame

   import "hiei-discord-bot/internal/interactions"

   func init() {
       router := interactions.GetRouter()
       router.RegisterComponent("game_yourgame_", handleComponent)
       router.RegisterModal("game_yourgame_modal_", handleModal)
   }
   ```
4. **Import in game.go**:
   ```go
   import _ "hiei-discord-bot/internal/commands/game/games/yourgame"
   ```
5. **Add subcommand** to `game.Definition()` in `game.go`
6. **Add case** to `game.Execute()` switch statement

The import triggers `init()` which registers handlers automatically.

### Adding Shared Functionality
If multiple commands need the same functionality:

1. **For configuration** - Add to `internal/config/config.go`
2. **For utilities** - Create new package in `internal/` (e.g., `internal/database/`)
3. **For bot features** - Add to `internal/bot/`
4. **For event handlers** - Add to `internal/events/`

Example shared database connection:
```go
// internal/database/database.go
package database

func Connect() (*DB, error) {
    // Connection logic
}

// In command:
import "hiei-discord-bot/internal/database"

func (c *Command) Execute(...) error {
    db, err := database.Connect()
    // Use db
}
```

## Troubleshooting

### Bot not responding to commands
- Verify `DISCORD_TOKEN` is correct in `.env`
- Check bot has "Message Content Intent" enabled in Discord Developer Portal
- Ensure bot has permissions in the Discord server

### Build errors
```bash
go mod tidy  # Clean up dependencies
go mod download  # Re-download dependencies
```

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]