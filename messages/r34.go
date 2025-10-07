package messages

import (
	"fmt"
	"io"
	"kannonfoundry/whutbot3/api"
	redgifsapi "kannonfoundry/whutbot3/api/redgifs"
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
	case "more":
		handleMoreCommand(s, m)
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

func handleMoreCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	msgs, err := s.ChannelMessages(m.ChannelID, 2, "", "", "")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}
	s.ChannelMessageSend(m.ChannelID, "More command received")
	for msg := range msgs {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Message: %s", msgs[msg].Content))
	}
}

func handleGimmeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	// Handle the "gimme" command
	var searchClient api.MediaSearcher
	var searchArgs string
	//check if first argument is gif
	command, gifArgs := parseCommand(args)
	if command == "gif" {
		searchClient = redgifsapi.NewClient()
		searchArgs = gifArgs
	} else {
		searchClient = rule34.NewClient()
		searchArgs = args
	}
	s.ChannelMessageSend(m.ChannelID, "Gimme command received with args: "+args)

	authorID, err := strconv.ParseInt(m.Author.ID, 10, 64)
	if err != nil {
		fmt.Printf("error parsing user ID: %v", err)
	}
	searchTerm, err := searchClient.FormatAndModifySearch(strings.Fields(searchArgs), authorID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error modifying search: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Gonna search for: "+searchTerm)
	files, err := searchClient.Search(strings.Fields(searchTerm))
	// Fetch posts from the Rule34 API
	// posts, err := rule34.GetPosts(strings.Fields(searchTerm))
	if err != nil {
		if err == io.EOF {
			s.ChannelMessageSend(m.ChannelID, "No posts found.")
		} else {
			fmt.Printf("error fetching posts: %v", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching posts: %v", err))
		}
		return
	}

	//check if the posts slice is empty

	if len(files) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No posts found.")
		return
	}
	//just get the first file for now
	req, err := http.NewRequest("GET", files[0].URL, nil)
	if err != nil {

	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = s.ChannelFileSend(m.ChannelID, files[0].Name, resp.Body)
	if err != nil {
		if strings.Contains(err.Error(), "entity too large") {
			s.ChannelMessageSend(m.ChannelID, "The booty too big 🥵")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error sending file: %v", err))
			fmt.Printf("error sending file: %v", err)
		}
	}

}
