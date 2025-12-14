# Hiei Discord Bot

A feature-rich Discord bot built with Go, featuring interactive games, multi-language support, and a modular command system.

## Features

- **Multi-language Support**: Automatic locale detection with support for Traditional Chinese (zh-TW) and English (en-US)
- **Interactive Games**: Bulls and Cows (1A2B) number guessing game with difficulty levels
- **Modular Architecture**: Clean, maintainable codebase with separation of concerns
- **Hot Reload**: Development mode with automatic restart on file changes
- **Guild-specific Commands**: Automatic command synchronization for each Discord server

## Commands

- `/ping` - Check bot responsiveness and latency
- `/help` - Display all available commands
- `/reload` - Reload slash commands for the current server (admin only)
- `/game bullsandcows [difficulty]` - Play the Bulls and Cows number guessing game
  - `easy` - Unique digits (no repeats)
  - `hard` - Repeating digits allowed

## Prerequisites

- [Go](https://golang.org/dl/) 1.25.5 or higher
- Discord Bot Token ([Create one here](https://discord.com/developers/applications))
- Discord Application ID

## Installation

1. Clone the repository:
```bash
git clone https://github.com/vongola12324/discord-bot.git
cd discord-bot
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file based on `.env.example`:
```bash
cp .env.example .env
```

4. Configure your `.env` file:
```env
DISCORD_TOKEN=your_bot_token_here
DISCORD_APP_ID=your_application_id_here
```

## Usage

### Running in Production

```bash
go run main.go
```

### Running in Development (Hot Reload)

Install Air for hot reload:
```bash
go install github.com/air-verse/air@latest
```

Start the bot with hot reload:
```bash
air
```

### Building

```bash
go build -o hiei-bot main.go
./hiei-bot
```

## Project Structure

```
discord-bot/
├── main.go                      # Application entry point
├── resources/                   # Embedded resources
│   ├── resources.go             # Resource embed declarations
│   └── i18n/                    # Translation files
│       ├── zh-TW.json           # Traditional Chinese
│       └── en-US.json           # English
├── internal/                    # Private application code
│   ├── bot/                     # Bot core logic
│   ├── commands/                # Slash command system
│   ├── interactions/            # Component & modal router
│   ├── events/                  # Discord event handlers
│   ├── i18n/                    # Internationalization
│   └── config/                  # Configuration management
├── .env.example                 # Environment template
└── .air.toml                    # Hot reload config
```

## Architecture

The bot uses a modular architecture with clear separation of concerns:

- **Commands**: Independent modules implementing the Command interface
- **Events**: Discord event handlers for bot lifecycle management
- **Interactions**: Router for button clicks and modal submissions
- **i18n**: Automatic locale detection with translation fallback

For detailed architecture documentation, see [CLAUDE.md](CLAUDE.md).

## Development

### Adding a New Command

1. Create a new directory under `internal/commands/`
2. Implement the `Command` interface
3. Register the command in `bot.registerCommands()`

Example:
```go
// internal/commands/mycommand/mycommand.go
package mycommand

import "github.com/bwmarrin/discordgo"

type Command struct{}

func New() *Command {
    return &Command{}
}

func (c *Command) Definition() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
        Name:        "mycommand",
        Description: "My command description",
    }
}

func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
    // Command logic here
    return nil
}
```

### Adding Translations

Add new keys to both `resources/i18n/zh-TW.json` and `resources/i18n/en-US.json`:

```json
{
  "mycommand": {
    "message": "Your translated message"
  }
}
```

Use in code:
```go
locale := i18n.GetUserLocaleFromInteraction(i)
message := i18n.T(locale, "mycommand.message")
```

### Testing

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord API wrapper for Go
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
- [Air](https://github.com/air-verse/air) - Live reload for Go apps

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/vongola12324/discord-bot/issues).
