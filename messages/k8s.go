package messages

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ParseCommand(input string) (command string, arguments string) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func HandleK8sMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	command, arguments := ParseCommand(m.Content)

	fmt.Printf("Command: '%s', Rest: '%s'", command, arguments)
	switch command {
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong")
	case "deployments":
		handleDeploymentsCommand(s, m, arguments)
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command")
	}
}
