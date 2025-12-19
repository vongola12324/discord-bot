package random

import (
	"fmt"
	"math/rand"
	"time"

	"hiei-discord-bot/internal/i18n"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

// Command implements the random command
type Command struct{}

// New creates a new random command instance
func New() *Command {
	return &Command{}
}

// Definition returns the slash command definition
func (c *Command) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "random",
		Description: "Generate random values",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "int",
				Description: "Generate a random integer",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "min",
						Description: "Minimum value (default: 0)",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "max",
						Description: "Maximum value (default: 100)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "string",
				Description: "Generate a random string",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "length",
						Description: "String length (default: 8)",
						Required:    false,
						MinValue:    float64Ptr(1),
						MaxValue:    2000,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "uuid",
				Description: "Generate a random UUID",
			},
		},
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// Version returns the command version
func (c *Command) Version() string {
	return "0.0.1"
}

// Execute runs the random command
func (c *Command) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	locale := i18n.GetUserLocaleFromInteraction(i)
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return nil
	}

	subcommand := data.Options[0]
	var result string

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch subcommand.Name {
	case "int":
		min := int64(0)
		max := int64(100)
		for _, opt := range subcommand.Options {
			if opt.Name == "min" {
				min = opt.IntValue()
			} else if opt.Name == "max" {
				max = opt.IntValue()
			}
		}
		if min > max {
			min, max = max, min
		}
		val := r.Int63n(max-min+1) + min
		result = fmt.Sprintf("%d", val)
	case "string":
		length := 8
		if len(subcommand.Options) > 0 && subcommand.Options[0].Name == "length" {
			length = int(subcommand.Options[0].IntValue())
		}
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[r.Intn(len(charset))]
		}
		result = string(b)
	case "uuid":
		result = uuid.New().String()
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: i18n.Tf(locale, "command.random.result", result),
		},
	})
}
