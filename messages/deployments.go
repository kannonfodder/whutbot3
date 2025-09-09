package messages

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func handleDeploymentsCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	command, arguments := parseCommand(args)
	switch command {
	case "help":
		s.ChannelMessageSend(m.ChannelID, "Available deployments commands: list [namespace], restart <namespace> <deployment>")
	case "list":
		handleDeploymentsListCommand(s, m, arguments)
	case "restart":
		handleDeploymentsRestartCommand(s, m, arguments)
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown deployments command")
	}
}

func handleDeploymentsListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {

	msg := "Listing all deployments..."
	if args != "" {
		msg += " in args"
	}

	s.ChannelMessageSend(m.ChannelID, msg)

	clientset, err := getClientSet(s, m)
	if err != nil {
		return
	}
	deps, err := clientset.AppsV1().Deployments(args).List(context.TODO(), metav1.ListOptions{})
	if k8sErrors.IsNotFound(err) {
		s.ChannelMessageSend(m.ChannelID, "Failed to get deployment - not found")
		fmt.Printf("%s", err.Error())
	} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus {
		fmt.Printf("%s", statusError.ErrStatus.Message)
	} else if err != nil {
		fmt.Printf("Error getting pod %v\n", err)
	} else {
		var depNames []string
		for _, dep := range deps.Items {
			depNames = append(depNames, fmt.Sprintf("%s - %s", dep.Namespace, dep.Name))
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Deployments found: namespace - name\n%s", strings.Join(depNames, "\n")))
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Failed to get deployment")

}

func handleDeploymentsRestartCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	clientset, err := getClientSet(s, m)
	if err != nil {
		return
	}
	if args == "" {
		s.ChannelMessageSend(m.ChannelID, "Please specify a namespace and deployment")
	}
	parts := strings.SplitN(args, " ", 3)
	if len(parts) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Please specify a namespace and deployment")
		return
	}
	namespace := parts[0]
	deployment := parts[1]
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Restarting deployment %s in namespace %s", deployment, namespace))
	_, err = clientset.AppsV1().Deployments(namespace).Patch(context.TODO(), deployment,
		types.JSONPatchType, []byte(`[{"op": "add", "path": "/spec/template/metadata/annotations/restartedAt", "value":"`+fmt.Sprintf("%d", metav1.Now().Unix())+`"}]`),
		metav1.PatchOptions{})
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to restart deployment")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Deployment restarted successfully")
}

func getClientSet(s *discordgo.Session, m *discordgo.MessageCreate) (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to create in-cluster config")
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to create k8s client")
		return nil, err
	}
	return clientset, nil
}
