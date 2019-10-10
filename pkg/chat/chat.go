package chat

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	chatv1 "github.com/knabben/chatops/api/v1"
	"github.com/nlopes/slack"
	cl "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type Chat struct {
	cl.Client
	SlackClient *slack.Client
	Log         logr.Logger
}

func NewChat(token string, client cl.Client) *Chat {
	return &Chat{SlackClient: slack.New(token), Client: client}
}

func (c *Chat) SendMessage(message string) {
	if _, _, err := c.SlackClient.PostMessage("CKY72UCH1", slack.MsgOptionText(message, false)); err != nil {
		c.Log.Error(err, "Error trying to send a message via slack.")
	}
}

func (c *Chat) UpdateItem(chat *chatv1.Chat, status chatv1.ChatStatus) {
	item := chat.DeepCopy()
	item.Spec.Timestamp = time.Now().Unix()
	item.Status = status

	if err := c.Client.Update(context.Background(), item); err != nil {
		fmt.Println("Update ERROR", err)
		return
	}
}

func (c *Chat) commandInCRD(command string) (*chatv1.Chat, error) {
	var chatList = &chatv1.ChatList{}
	if err := c.List(context.Background(), chatList, &cl.ListOptions{}); err != nil {
		fmt.Println("List ERROR", err)
		return nil, err
	}

	for _, item := range chatList.Items {
		if item.Spec.Command == command {
			return &item, nil
		}
	}

	return nil, nil
}

// ListenChat listen for events in the chat channel
func (c *Chat) ListenChat(inputChannel chan *chatv1.ChatStatus) {
	rtm := c.SlackClient.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			chatStatus := c.ExtractChatStatus(ev)
			if ev.Username != "td" && chatStatus != nil { // filter commands here before listing on CRD?
				fmt.Println(fmt.Sprintf("Message: %v, %s\n", ev.Text, ev.User))
				inputChannel <- chatStatus
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		}
	}
}

// ChangeCRD updates the resource definition based on Slack events
func (c *Chat) ChangeCRD(inputChannel chan *chatv1.ChatStatus) {
	for {
		chatStatus := <-inputChannel
		if chat, _ := c.commandInCRD(chatStatus.Command); chat != nil {
			c.UpdateItem(chat, *chatStatus)
		}
	}
}

// ExtractChatStatus returns a translated content of chat status
func (c *Chat) ExtractChatStatus(message *slack.MessageEvent) *chatv1.ChatStatus {
	tokens := strings.Split(message.Text, " ")
	arguments := strings.Join(tokens[1:], " ")

	if len(tokens) < 1 || len(arguments) == 0  {
		return nil
	}

	return &chatv1.ChatStatus{
		Command:   tokens[0],
		Arguments: arguments,
		Username:  message.Username,
		Timestamp: time.Now().String(),
		Channel:   message.Channel,
	}
}
