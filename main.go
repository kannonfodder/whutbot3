package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"kannonfoundry/whutbot3/config"
	"kannonfoundry/whutbot3/dotenv"
	"kannonfoundry/whutbot3/messages"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Load .env if present (values do not override existing environment variables)
	dotenv.Load(".env")

	cfg := config.Default()
	token := cfg.Token

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	// Request the guild message and message content intents so we can read messages
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		messages.DispatchMessageByChannel(messages.DefaultHandlers(cfg))(s, m)
	})

	dg.ChannelMessageSend(cfg.LogChannelID, "WhutBot is now running and listening")

	if err = dg.Open(); err != nil {
		dg.ChannelMessageSend(cfg.LogChannelID, "WhutBot is shutting down.")
		log.Fatalf("error opening connection: %v", err)
	}
	defer dg.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	dg.ChannelMessageSend(cfg.LogChannelID, "WhutBot is shutting down.")
	log.Println("Shutting down.")
}
