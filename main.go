package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/whutbot3/dotenv"
	"example.com/whutbot3/messages"
	"github.com/bwmarrin/discordgo"
)

func main() {
	// Load .env if present (values do not override existing environment variables)
	dotenv.Load(".env")

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN not set")
	}
	channelID := os.Getenv("TARGET_CHANNEL_ID")
	if channelID == "" {
		log.Fatal("TARGET_CHANNEL_ID not set")
	}

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
		if m.ChannelID != channelID {
			return
		}
		// delegate message processing to the messages package
		messages.HandleMessage(s, m)
	})

	dg.ChannelMessageSend(channelID, "WhutBot is now running and listening for stashdb links.")

	if err = dg.Open(); err != nil {
		dg.ChannelMessageSend(channelID, "WhutBot is shutting down.")
		log.Fatalf("error opening connection: %v", err)
	}
	defer dg.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	dg.ChannelMessageSend(channelID, "WhutBot is shutting down.")
	log.Println("Shutting down.")
}
