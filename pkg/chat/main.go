package chat

import (
	"fmt"
	"github.com/nlopes/slack"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ChangeCRD(inputChannel, outputChannel chan string, client client.Client) {
	for {
		switch <- inputChannel {

		case "bla":
			// Change CRD value
			fmt.Println("OUTPUT CHANNEL CONSUMER - ", <- outputChannel)
		}
	}
}
func ListenChat(inputChannel chan string) {
	api := slack.New("")

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