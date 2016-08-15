package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"time"
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
	Long:  `Consume messages from AMQP and start Kubernetes Jobs`,
	Run: func(cmd *cobra.Command, args []string) {
		QueueConsumer()
		select {}
	},
}

func initRabbitConn(consumer Consumer) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				var err error
				consumer.conn, err = amqp.Dial(Uri)
				if err != nil {
					fmt.Println(err)
					fmt.Println("node will only be able to respond to local connections")
					fmt.Println("trying to reconnect in 5 seconds...")
				} else {
					close(quit)

					// Create a channel and declare a queue to consume from
					consumer.channel, err = consumer.conn.Channel()
					queue, err := consumer.channel.QueueDeclare(
						"runner", true, false, false, false, nil,
					)
					if err != nil {
						fmt.Println("ERROR: on runner queue declare")
					}
					deliveries, err := consumer.channel.Consume(
						queue.Name, "", false, false, false, false, nil,
					)
					if err != nil {
						fmt.Println("ERROR: trying to consume runner")
					}
					go handle(deliveries, consumer)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func QueueConsumer() {
	consumer := &Consumer{
		conn:    nil,
		channel: nil,
		done:    make(chan error),
	}
	initRabbitConn(*consumer)
}

func decodeBody(body []byte) (e Event) {
	var event Event
	err := json.Unmarshal(body, &event)
	if err != nil {
		fmt.Println("ERROR: decoding body - %s", err)
	}
	return event
}

func handle(deliveries <-chan amqp.Delivery, consumer Consumer) {
	var podName string = "script"

	ca, err := ioutil.ReadFile(KubeCA)
	if err != nil {
		fmt.Println("ERROR: opening the CA file", err)
	}
	token, err := ioutil.ReadFile(KubeToken)
	if err != nil {
		fmt.Println("ERROR: opening the TOKEN file")
	}

	config := &restclient.Config{
		Host:            KubeHost,
		TLSClientConfig: restclient.TLSClientConfig{CAData: ca},
		BearerToken:     string(token[:]),
	}
	kubeClient, err := client.New(config)
	if err != nil {
		fmt.Println("ERROR: Can't connect on Kubernetes server ", err)
	}

	// Consume messages and run the pods
	for msg := range deliveries {
		body := decodeBody(msg.Body)
		createPod(kubeClient, body, podName)
		ReadLogAndPublish(kubeClient, consumer, podName)
		deletePod(kubeClient, podName)
		msg.Ack(true)
	}
	consumer.done <- nil
}

func init() {
	RootCmd.AddCommand(consumeCmd)
	consumeCmd.Flags().StringVar(&Uri, "uri", "amqp://guest:guest@localhost:5672", "AQMP default URI")
	consumeCmd.Flags().StringVar(&KubeHost, "kubehost", "https://lo", "Kubernetes host")
	consumeCmd.Flags().StringVar(&KubeCA, "kubeca", "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt", "Kubernetes CA")
	consumeCmd.Flags().StringVar(&KubeToken, "kubetok", "/var/run/secrets/kubernetes.io/serviceaccount/token", "Kubernetes Token")
}
