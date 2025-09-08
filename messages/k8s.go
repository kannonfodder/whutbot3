package messages

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func HandleK8sMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}
	firstIdx := strings.Index(m.Content, " ")
	command := m.Content[:firstIdx]
	rest := m.Content[firstIdx+1:]

	fmt.Printf("Command: '%s', Rest: '%s'", command, rest)
	switch command {
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong")
	case "get":
		handleGetCommand(s, m, rest)
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command")
	}
}

func handleGetCommand(s *discordgo.Session, m *discordgo.MessageCreate, restOfCommand string) {
	if restOfCommand == "deployments" {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to create in-cluster config")
			return
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to create k8s client")
			return
		}
		dep, err := clientset.AppsV1().Deployments("bot").Get(context.TODO(), restOfCommand, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) {
			fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			fmt.Printf("Error getting pod %v\n", err)
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Deployment found: %s", dep.Name))
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Failed to get deployment")
		return
	} else {
		s.ChannelMessageSend(m.ChannelID, "Unknown resource")
	}
}
