package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/whutbot3/dotenv"
	"example.com/whutbot3/messages"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token             string
	WhisparrChannelID string
	K8SChannelID      string
	LogChannelID      string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		Token:             os.Getenv("DISCORD_TOKEN"),
		WhisparrChannelID: os.Getenv("WHISPARR_CHANNEL_ID"),
		K8SChannelID:      os.Getenv("K8S_CHANNEL_ID"),
		LogChannelID:      os.Getenv("LOG_CHANNEL_ID"),
	}
	newError := errors.New("config error")
	errString := ""
	if cfg.Token == "" {
		errString += fmt.Sprintf("%w: DISCORD_TOKEN\n", newError)
	}
	if cfg.WhisparrChannelID == "" {
		errString += fmt.Sprintf("%w: WHISPARR_CHANNEL_ID\n", newError)
	}
	if cfg.K8SChannelID == "" {
		errString += fmt.Sprintf("%w: K8S_CHANNEL_ID\n", newError)
	}
	if cfg.LogChannelID == "" {
		errString += fmt.Sprintf("%w: ERROR_CHANNEL_ID\n", newError)
	}
	if errString != "" {
		return nil, fmt.Errorf("%w: %s", newError, errString)
	}
	return cfg, nil
}

func main() {
	// Load .env if present (values do not override existing environment variables)
	dotenv.Load(".env")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

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
		messages.DispatchMessageByChannel(map[string]messages.HandlerFunc{
			cfg.WhisparrChannelID: messages.HandleStashMessage,
			cfg.K8SChannelID:      messages.HandleK8sMessage,
		})(s, m)
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
