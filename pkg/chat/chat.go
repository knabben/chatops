package chat

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	chatv1 "github.com/knabben/chatops/api/v1"
	"github.com/nlopes/slack"
	"k8s.io/apimachinery/pkg/types"
	cl "sigs.k8s.io/controller-runtime/pkg/client"
	//"strings"
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

func (c *Chat) commandInCRD(command string) *chatv1.Chat {
	var chat chatv1.Chat
	objectKey := types.NamespacedName{
		Namespace: "default",
		Name:      "chat-sample",
	}

	err := c.Get(context.Background(), objectKey, &chat)
	if err != nil {
		return nil
	}
	return &chat
}

// ChangeCRD updates the resource definition based on Slack events
func (c *Chat) ChangeCRD(inputChannel chan *chatv1.ChatStatus) {
	for {
		chatStatus := <-inputChannel
		if chat := c.commandInCRD(chatStatus.Command); chat != nil {
			c.UpdateItem(chat, *chatStatus)
		}
	}
}

// ListenChat listen for events in the chat channel
func (c *Chat) ListenChat(inputChannel chan *chatv1.ChatStatus) {
	rtm := c.SlackClient.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if c.FilterValidMessage(ev) {
				fmt.Println(fmt.Sprintf("Message: %v, %s\n", ev.Text, ev.User))
				chatStatus := c.ExtractChatStatus(ev)
				inputChannel <- chatStatus
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		}
	}
}

// FilterValidMessage filters ConfigMaps to match possible regex and ACL
func (c *Chat) FilterValidMessage(message *slack.MessageEvent) bool {
	return message.SubType != "bot_message"
}

// ExtractChatStatus returns a translated content of chat status
func (c *Chat) ExtractChatStatus(message *slack.MessageEvent) *chatv1.ChatStatus {
	return &chatv1.ChatStatus{
		Command:   message.Text,
		Username:  message.User,
		Timestamp: time.Now().String(),
		Channel:   message.Channel,
	}
}
