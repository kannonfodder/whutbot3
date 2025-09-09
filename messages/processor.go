package messages

import (
	"kannonfoundry/whutbot3/config"

	"github.com/bwmarrin/discordgo"
)

// HandlerFunc defines the signature for message handler functions.
type HandlerFunc func(s *discordgo.Session, m *discordgo.MessageCreate)

func DefaultHandlers(cfg *config.Config) map[string]HandlerFunc {
	return map[string]HandlerFunc{
		cfg.WhisparrChannelID: HandleStashMessage,
		cfg.K8SChannelID:      HandleK8sMessage,
		cfg.R34ChannelID:      HandleR34Message,
	}
}

// DispatchMessageByChannel dispatches message handling based on channel ID.
// handlers is a map of channel IDs to handler functions.
func DispatchMessageByChannel(handlers map[string]HandlerFunc) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		handler, ok := handlers[m.ChannelID]
		if ok {
			handler(s, m)
		}
	}
}
