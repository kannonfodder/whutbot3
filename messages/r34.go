package messages

import (
	"fmt"
	"io"
	"kannonfoundry/whutbot3/api"
	redgifsapi "kannonfoundry/whutbot3/api/redgifs"
	"kannonfoundry/whutbot3/api/rule34"
	prefs "kannonfoundry/whutbot3/db/preferences"
	"kannonfoundry/whutbot3/db/sent"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		handleMoreCommand(s, m, 0, "")
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
		err = prefs.SetPreferences(authorID, strings.Split(arguments, " "))
	case "add":
		err = prefs.AddPreferences(authorID, strings.Split(arguments, " "))
	case "remove":
		err = prefs.RemovePreferences(authorID, strings.Split(arguments, " "))

	case "list":
		prefs, err := prefs.GetPreferences(authorID)
		if err == nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Your preferences: %s", prefs.String()))
		}
	}
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error handling preferences: %v", err))
	}
}

func handleMoreCommand(s *discordgo.Session, m *discordgo.MessageCreate, depth int, lastMessageID string) {
	if depth == 0 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔍")
		s.ChannelMessageSend(m.ChannelID, "More command received")
	}
	found := false
	msgs, err := s.ChannelMessages(m.ChannelID, 5, lastMessageID, "", "")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	for _, msg := range msgs {
		if msg.Author.Bot {
			continue
		}
		if strings.HasPrefix(msg.Content, "gimme") {
			found = true
			_, arguments := parseCommand(msg.Content)
			handleGimmeCommand(s, m, arguments)
			s.MessageReactionRemove(m.ChannelID, m.ID, "🔍", s.State.User.ID)
		}

	}
	if !found && depth < 3 {
		handleMoreCommand(s, m, depth+1, msgs[len(msgs)-1].ID)
	} else if !found {
		s.MessageReactionRemove(m.ChannelID, m.ID, "🔍", s.State.User.ID)
		s.ChannelMessageSend(m.ChannelID, "No recent gimme command found.")
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
	//s.ChannelMessageSend(m.ChannelID, "Gimme command received with args: "+args)
	fmt.Printf("Searching for: %v", searchArgs)
	authorID, err := strconv.ParseInt(m.Author.ID, 10, 64)
	if err != nil {
		fmt.Printf("error parsing user ID: %v", err)
	}
	searchTerm, err := searchClient.FormatAndModifySearch(strings.Fields(searchArgs), authorID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error modifying search: %v", err))
		return
	}

	searchMsg, _ := s.ChannelMessageSend(m.ChannelID, "Gonna search for: "+searchTerm)
	// Fetch posts from the API
	files, err := searchClient.Search(strings.Fields(searchTerm))
	if err != nil {
		if err == io.EOF {
			s.ChannelMessageSend(m.ChannelID, "No posts found.")
		} else {
			fmt.Printf("error fetching posts: %v", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching posts: %v", err))
		}
		return
	}

	fmt.Printf("Found files: %d", len(files))
	//check if the posts slice is empty

	if len(files) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No posts found.")
		return
	}
	//just get the first file for now

	sentDB, err := sent.NewSentDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error initializing sent database: %v", err))
		return
	}
	defer sentDB.Close()
	var fileUrl = ""
	for _, file := range files {
		beenSent, err := sentDB.HasBeenSent(file.URL)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error checking sent database: %v", err))
			return
		}
		if !beenSent {
			fileUrl = file.URL
		}
	}
	if fileUrl == "" {
		fileUrl = files[0].URL
		s.ChannelMessageSend(m.ChannelID, "No new files found")
	}
	req, err := http.NewRequest("GET", files[0].URL, nil)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error creating HTTP request: %v", err))
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v, %v", fileUrl, time.Now().UnixMilli()))
	err = sentDB.MarkAsSent(fileUrl)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error marking post as sent: %v", err))
		return
	}

	_, err = s.ChannelFileSend(m.ChannelID, files[0].Name, resp.Body)
	if err != nil {
		if strings.Contains(err.Error(), "entity too large") {
			s.ChannelMessageSend(m.ChannelID, "The booty too big 🥵")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error sending file: %v", err))
			fmt.Printf("error sending file: %v", err)
		}
	} else {
		s.ChannelMessageDelete(searchMsg.ChannelID, searchMsg.ID)
	}

}
