package chat

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
	"os"
	c "sigs.k8s.io/controller-runtime/pkg/client"
	chatv1alpha1 "github.com/knabben/chatops/pkg/apis/chat/v1alpha1"
	"time"
)

var token = os.Getenv("TOKEN")

func ChangeCRD(inputChannel chan *chatv1alpha1.Chat, outputChannel chan string, client c.Client) {
	api := slack.New(token)

	for {
		chat := <- inputChannel
		switch chat.Status.Command {
		case "chat":
			// Change CRD value
			instance := &chatv1alpha1.ChatList{}
			client.List(context.TODO(), nil, instance)

			for _, item := range instance.Items {
				chatItem := item.DeepCopy()

				chatItem.Status = chat.Status
				err := client.Status().Update(context.Background(), chatItem)
				if err != nil {
					fmt.Println(err)
					continue
				}

				chatItem.Spec.Timestamp = time.Now().Unix()
				err = client.Update(context.Background(), chatItem)
				if err != nil {
					fmt.Println(err)
					continue
				}

			}
			_, _, err := api.PostMessage("CKY72UCH1", slack.MsgOptionText("ok", false))
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("-------")
		}
	}
}

func ListenChat(inputChannel chan *chatv1alpha1.Chat) {
	api := slack.New(token)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if ev.User != "" {
				fmt.Printf("Message: %v, %s\n", ev.Text, ev.User)
				inputChannel <- &chatv1alpha1.Chat{
					Status: chatv1alpha1.ChatStatus{
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