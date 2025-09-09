package messages

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func HandleK8sMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	command, arguments := parseCommand(m.Content)

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
