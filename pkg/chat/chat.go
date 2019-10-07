package chat

import (
	"context"
	"fmt"
	chatv1 "github.com/knabben/chatops/api/v1"
	"github.com/nlopes/slack"
	"os"
	c "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	token = os.Getenv("TOKEN")
)

func UpdateItems(chatList *chatv1.ChatList, chatStatus *chatv1.Chat, client c.Client) {
	for _, item := range chatList.Items {
		chatItem := item.DeepCopy()
		chatItem.Spec.Timestamp = time.Now().Unix()
		if err := client.Update(context.Background(), chatItem); err != nil {
			fmt.Println("UpdateERROR", err)
			continue
		}
	}
}

func ChangeCRD(inputChannel chan *chatv1.Chat, outputChannel chan string, client c.Client) {
	api := slack.New(token)

	for {
		chat := <-inputChannel
		switch chat.Status.Command {
		case "chat":
			var chatList = &chatv1.ChatList{}
			if err := client.List(context.Background(), chatList, &c.ListOptions{}); err != nil {
				fmt.Println("List error", err)
				continue
			}

			UpdateItems(chatList, chat, client)

			if _, _, err := api.PostMessage("CKY72UCH1", slack.MsgOptionText("ok", false)); err != nil {
				fmt.Println(err)
			}
			fmt.Println("-------")
		}
	}
}

func ListenChat(inputChannel chan *chatv1.Chat) {
	api := slack.New(token)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if ev.User != "" {
				fmt.Printf("Message: %v, %s\n", ev.Text, ev.User)
				inputChannel <- &chatv1.Chat{
					Status: chatv1.ChatStatus{
						Command:  ev.Text,
						Username: ev.User,
					},
				}
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		}
	}
}
