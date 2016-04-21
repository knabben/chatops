package main

import (
	"flag"
	"fmt"
	"log"

	"encoding/json"

	"github.com/streadway/amqp"

	"github.com/fsouza/go-dockerclient"
)

var (
	uri      = flag.String("uri", "amqp://guest:guest@192.168.0.6:5672/", "AMQP URI")
	endpoint = flag.String("endpoint", "tcp://192.168.0.6:2375", "DOCKER HOST")
	queue    = flag.String("queue", "runner", "AMQP Queue")
)

type Event struct {
	Cmd    string `json:"command"`
	Params string `json:"params"`
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
}

func main() {
	flag.Parse()

	// start hubot queue dispatcher
	StartQueueConsumer(*uri, *queue)
	select {}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func StartQueueConsumer(amqpURI string, queueName string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		done:    make(chan error),
	}

	var err error

	log.Printf("AMQP Connection on: %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	failOnError(err, "Error loading connection")

	go func() {
		fmt.Printf("Closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	c.channel, err = c.conn.Channel()
	failOnError(err, "Error creating Channel")
	log.Printf("Starting Channel")

	q, err := c.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	log.Printf("Declare Queue: %s", q.Name)

	deliveries, err := c.channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed trying to consuming the queue")

	go handle(deliveries, c.done)

	return c, nil
}

func ErrorOnResponseQueue(msg string) {
	fmt.Println(msg)
}

func (e Event) decodeBody(body []byte) {
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println("ERROR %s", err)
	}
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	client, _ := docker.NewClient(*endpoint)

	for d := range deliveries {
		found := false

		containers, _ := client.ListContainers(docker.ListContainersOptions{})
		for _, c := range containers {
			if c.Names[0] == "/worker" {
				// We found our worker
				err := client.StartContainer("worker", &docker.HostConfig{})
				if err != nil {
					ErrorOnResponseQueue(err.Error())
				}
				found = true
			}
		}

		if !found {
			ErrorOnResponseQueue("Worker container not found, try to create it before starting!")
		}
		d.Ack(true)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
