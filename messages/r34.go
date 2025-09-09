package messages

import (
	"fmt"
	"kannonfoundry/whutbot3/api/rule34"
	"kannonfoundry/whutbot3/db"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func HandleR34Message(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	command, arguments := parseCommand(m.Content)
	switch command {
	case "prefs":
		handlePrefsCommand(s, m, arguments)
	case "gimme":
		handleGimmeCommand(s, m, arguments)
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown r34 command")
	}
}

func handlePrefsCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	// Handle the "prefs" command
	command, arguments := parseCommand(args)
	authorID, err := strconv.ParseInt(m.Author.ID, 10, 64)
	if err != nil {
		fmt.Printf("error parsing user ID: %v", err)
	}
	switch command {
	case "set":
		err = db.SetPreferences(authorID, strings.Split(arguments, " "))
	case "add":
		err = db.AddPreferences(authorID, strings.Split(arguments, " "))
	case "remove":
		err = db.RemovePreferences(authorID, strings.Split(arguments, " "))

	case "list":
		prefs, err := db.GetPreferences(authorID)
		if err == nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Your preferences: %s", prefs.String()))
		}
	}
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error handling preferences: %v", err))
	}
}

func handleGimmeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	// Handle the "gimme" command

	authorID, err := strconv.ParseInt(m.Author.ID, 10, 64)
	if err != nil {
		fmt.Printf("error parsing user ID: %v", err)
	}
	prefs, err := db.GetPreferences(authorID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error retrieving preferences: %v", err))
		return
	}

	searchTerm := args
	for _, pref := range prefs {
		searchTerm += " " + pref.Preference
	}

	s.ChannelMessageSend(m.ChannelID, "Gimme command received with args: "+args)
	s.ChannelMessageSend(m.ChannelID, "Gonna search for: "+searchTerm)

	// Fetch posts from the Rule34 API
	posts, err := rule34.GetPosts(strings.Fields(searchTerm))
	if err != nil {
		fmt.Printf("error fetching posts: %v", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching posts: %v", err))
		return
	}

	//check if the posts slice is empty
	if len(posts) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No posts found.")
		return
	}
	req, err := http.NewRequest("GET", posts[0].FileURL, nil)
	if err != nil {

	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = s.ChannelFileSend(m.ChannelID, posts[0].FileName, resp.Body)
	if err != nil {
		if strings.Contains(err.Error(), "entity too large") {
			s.ChannelMessageSend(m.ChannelID, "The booty too big 🥵")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error sending file: %v", err))
			fmt.Printf("error sending file: %v", err)
		}
	}

}
