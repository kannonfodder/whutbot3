package messages

import (
	"errors"
	"log"
	"strings"

	"example.com/whutbot3/api"
	"github.com/bwmarrin/discordgo"
)

// StashPrefix is the prefix we look for in messages.
const StashPrefix = "https://stashdb.org/scenes"

// ParseStashLink extracts the scene identifier from a stashdb scenes URL.
// Examples it accepts:
// - https://stashdb.org/scenes/12345
// - https://stashdb.org/scenes/12345/whatever
// - https://stashdb.org/scenes/12345?foo=bar
func ParseStashLink(url string) (string, error) {
	if !strings.HasPrefix(url, StashPrefix) {
		return "", errors.New("not a stashdb scenes url")
	}
	rest := strings.TrimPrefix(url, StashPrefix)
	rest = strings.TrimPrefix(rest, "/")
	if rest == "" {
		return "", errors.New("no scene id in url")
	}
	// take up to the next slash or query
	end := len(rest)
	if i := strings.IndexAny(rest, "/?"); i >= 0 {
		end = i
	}
	scene := rest[:end]
	if scene == "" {
		return "", errors.New("empty scene id")
	}
	return scene, nil
}

// HandleMessage processes a Discord message and responds when it contains a stashdb link.
// It does not filter by channel — the caller should ensure channel filtering if desired.
func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	// find the first occurrence of the stash prefix in the message content
	idx := strings.Index(m.Content, StashPrefix)
	if idx == -1 {
		return
	}
	// extract the candidate URL (up to next whitespace)
	tail := m.Content[idx:]
	fields := strings.Fields(tail)
	if len(fields) == 0 {
		return
	}
	url := fields[0]

	sceneID, err := ParseStashLink(url)
	if err != nil {
		// not a parseable link — ignore
		return
	}

	log.Printf("matched stashdb link from %s: %s (scene=%s)", m.Author.Username, url, sceneID)

	// acknowledge in-channel
	_, err = s.ChannelMessageSend(m.ChannelID, "Received stashdb link — processing...")
	if err != nil {
		log.Printf("failed to send reply: %v", err)
	}
	// add a peach reaction to the original message
	if err := s.MessageReactionAdd(m.ChannelID, m.ID, "👀"); err != nil {
		log.Printf("failed to add reaction: %v", err)
	}
	// check existence with Whispar
	exists, err := api.LookupScene(sceneID)
	if err != nil {
		log.Printf("whispar lookup error for scene %s: %v", sceneID, err)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error checking scene existence.")
		return
	}
	if !exists {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Scene not found in Whispar.")
		if success, err := api.AddScene(sceneID); success {
			if err := s.MessageReactionAdd(m.ChannelID, m.ID, "🍑"); err != nil {
				log.Printf("failed to add reaction: %v", err)
			}
			if err := s.MessageReactionRemove(m.ChannelID, m.ID, "👀", "@me"); err != nil {
				log.Printf("failed to remove reaction: %v", err)
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "Added scene to Whispar.")
		} else {
			log.Printf("failed to add scene %s: %v", sceneID, err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "Failed to add scene to Whispar.")
		}
		return
	}
	if exists {
		log.Printf("scene %s exists in Whispar", sceneID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Scene already in Whispar.")
		if err := s.MessageReactionAdd(m.ChannelID, m.ID, "🍑"); err != nil {
			log.Printf("failed to add reaction: %v", err)
		}
		if err := s.MessageReactionRemove(m.ChannelID, m.ID, "👀", "@me"); err != nil {
			log.Printf("failed to remove reaction: %v", err)
		}

	}
}
