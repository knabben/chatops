package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var Uri, Kubeconfig, Varpath string

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
}

// Event catcher
type Event struct {
	Args  []string `json:"args"`
	Image string   `json:"image"`
	Pod   string   `json:"pod"`
	Id    string   `json:"id"`
}

// Serialized EnvVar struct from Yaml
type varEnv struct {
	Name   string
	Key    string
	Secret string
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
	ticker := time.NewTicker(20 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				var err error
				consumer.conn, err = amqp.Dial(Uri)
				if err != nil {
					fmt.Println(err)
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
					// Gorotine to consumer messages from request queue
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
		log.Fatal(err)
	}
	return event
}

func handle(deliveries <-chan amqp.Delivery, consumer Consumer) {
	log.Println("Connecting on Kubernetes")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println(`You're not in a cluster, trying the local configuration,
					if not possible it should die on next try.`)
		config, err = clientcmd.BuildConfigFromFlags("", Kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	log.Println("Cool, Kubernetes konnected!")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	data, err := ioutil.ReadFile(Varpath)

	// All variable env should be stored on a Secret object
	envList := []varEnv{}
	err = yaml.Unmarshal(data, &envList)
	if err != nil {
		log.Fatal("ERROR trying to parse variables YAML", err)
	}

	log.Println("Starting listening on broker consumer channel")
	for msg := range deliveries {
		// Consume messages and run the pods
		go func() {
			event := decodeBody(msg.Body)

			// StartPod
			pod := createPod(clientset, event, msg, envList)
			ReadLogAndPublish(clientset, consumer, pod)

			// Remove pod
			deletePod(clientset, event.Id)
		}()
		msg.Ack(true)
	}
	consumer.done <- nil
}

func init() {
	RootCmd.AddCommand(consumeCmd)
	consumeCmd.Flags().StringVar(&Uri, "uri", "amqp://guest:guest@rabbitmq:5672", "AQMP default URI")
	consumeCmd.Flags().StringVar(&Kubeconfig, "kubeconfig", "~/.kube/config", "Kubeconfig path")
	consumeCmd.Flags().StringVar(&Varpath, "var", "var.yaml", "Variables environment (Secrets) for pods")
}
