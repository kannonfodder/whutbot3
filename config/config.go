package config

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Token             string
	WhisparrChannelID string
	K8SChannelID      string
	LogChannelID      string
	R34ChannelID      string
	R34ApiKey         string
	R34UserID         string
}

func Default() *Config {
	cfg := &Config{
		Token:             os.Getenv("DISCORD_TOKEN"),
		WhisparrChannelID: os.Getenv("WHISPARR_CHANNEL_ID"),
		K8SChannelID:      os.Getenv("K8S_CHANNEL_ID"),
		LogChannelID:      os.Getenv("LOG_CHANNEL_ID"),
		R34ChannelID:      os.Getenv("R34_CHANNEL_ID"),
		R34ApiKey:         os.Getenv("R34_API_KEY"),
		R34UserID:         os.Getenv("R34_USER_ID"),
	}
	newError := errors.New("config error")
	errString := ""
	if cfg.Token == "" {
		errString += fmt.Sprintf("%s: DISCORD_TOKEN\n", newError)
	}
	if cfg.WhisparrChannelID == "" {
		errString += fmt.Sprintf("%s: WHISPARR_CHANNEL_ID\n", newError)
	}
	if cfg.K8SChannelID == "" {
		errString += fmt.Sprintf("%s: K8S_CHANNEL_ID\n", newError)
	}
	if cfg.LogChannelID == "" {
		errString += fmt.Sprintf("%s: ERROR_CHANNEL_ID\n", newError)
	}
	if cfg.R34ChannelID == "" {
		errString += fmt.Sprintf("%s: R34_CHANNEL_ID\n", newError)
	}
	if cfg.R34ApiKey == "" {
		errString += fmt.Sprintf("%s: R34_API_KEY\n", newError)
	}
	if cfg.R34UserID == "" {
		errString += fmt.Sprintf("%s: R34_USER_ID\n", newError)
	}
	if errString != "" {
		log.Fatalf("error loading config: %v", errString)
	}
	return cfg
}
