package chat

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
	"math/rand"
	c"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	chatv1alpha1 "github.com/knabben/chatops/pkg/apis/chat/v1alpha1"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func NewString(length int) string {
	return StringWithCharset(length, charset)
}

var token = "YOUR_SLACK_TOKEN_HERE"

func ChangeCRD(inputChannel, outputChannel chan string, client c.Client) {
	api := slack.New(token)

	for {
		switch <- inputChannel {

		case "bla":
			// Change CRD value
			instance := &chatv1alpha1.ChatList{}
			client.List(context.TODO(), nil, instance)

			for _, chat := range instance.Items {
				chat1 := chat.DeepCopy()
				chat1.Spec.Halo = NewString(10)
				fmt.Println("DEBUGGING - chat1", chat1)
				err := client.Update(context.Background(), chat1)
				fmt.Println(err)
			}
			fmt.Println("-------")
			_,_ , err := api.PostMessage("CKY72UCH1", slack.MsgOptionText(<-outputChannel, false))
			fmt.Println(err)
			fmt.Println("-------")
		}
	}
}
func ListenChat(inputChannel chan string) {
	api := slack.New(token)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("\nEvent Received: \n", msg.Data)

		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			inputChannel <- ev.Text

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		}
	}
}