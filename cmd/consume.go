package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"

)

var Uri string
var KubeHost string
var KubeCA string
var KubeToken string

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
}

type Event struct {
	Cmd    string `json:"command"`
	Params string `json:"params"`
}

var consumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "Consume messages from AMQP and start Kubernetes Jobs",
	Long: `Consume messages from AMQP and start Kubernetes Jobs`,
	Run: func(cmd *cobra.Command, args []string) {
		QueueConsumer()
		select {}
	},
}

func QueueConsumer() {
	c := &Consumer{
		conn: nil,
		channel: nil,
		done: make(chan error),
	}
	var err error

	// Start connection
	c.conn, err = amqp.Dial(Uri)
	if err != nil {
		fmt.Printf("Dial: %s", err)
	}

	// Get channel from connection
	c.channel, err = c.conn.Channel()
	queue, err := c.channel.QueueDeclare(
		"runner", true, false, false, false, nil,
	)
	if err != nil {
		fmt.Println("No queue declare")
	}

	// Consume it as go channel
	deliveries, err := c.channel.Consume(
		queue.Name, "", false, false, false, false, nil,
	)
	go handle(deliveries, *c)
}


func decodeBody(body []byte) (e Event) {
	var event Event
	err := json.Unmarshal(body, &event)
	if err != nil {
		fmt.Println("ERROR %s", err)
	}
	return event
}


func handle(deliveries <-chan amqp.Delivery, c Consumer) {
	ca, err := ioutil.ReadFile(KubeCA)
	if err != nil {
		fmt.Println("Error opening the CA file", err)
	}
	token, err := ioutil.ReadFile(KubeToken)
	if err != nil {
		fmt.Println("Error opening the TOKEN file")
	}

	config := &restclient.Config{
		Host: KubeHost,
		TLSClientConfig: restclient.TLSClientConfig{CAData: ca},
		BearerToken: string(token[:]),
	}
	kubeClient, err := client.New(config)
	if err != nil {
		fmt.Println("Can't connect on Kubernetes server: ", err)
	}

	for d := range deliveries {
		createJob(kubeClient)
		e := decodeBody(d.Body)
		fmt.Println(e.Cmd)
		d.Ack(true)
	}
	c.done <- nil
}

func init() {
	RootCmd.AddCommand(consumeCmd)
	consumeCmd.Flags().StringVar(&Uri, "uri", "amqp://guest:guest@localhost:5672", "AQMP default URI")
	consumeCmd.Flags().StringVar(&KubeHost, "kubehost", "https://lo", "Kubernetes host")
	consumeCmd.Flags().StringVar(&KubeCA, "kubeca", "/home/ubuntu/.kube/credentials/ca.pem", "Kubernetes CA")
	consumeCmd.Flags().StringVar(&KubeToken, "kubetok", "/home/ubuntu/.kube/credentials/token", "Kubernetes Token")
}
