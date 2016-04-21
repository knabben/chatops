package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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

	go handle(deliveries, *c)

	return c, nil
}

func ErrorOnResponseQueue(msg string) {
	fmt.Println(msg)
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
	client, _ := docker.NewClient(*endpoint)

	for d := range deliveries {
		e := decodeBody(d.Body)

		// Removing old container
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    e.Cmd,
			Force: true,
		})

		host_config := docker.HostConfig{}
		opts := docker.CreateContainerOptions{
			Name: e.Cmd,
			Config: &docker.Config{
				Image: "debian",
				Cmd:   []string{e.Params},
			},
			HostConfig: &host_config,
		}
		// Creating a new container
		cont, _ := client.CreateContainer(opts)

		// Starting the container
		client.StartContainer(cont.Name, &host_config)

		reader, writer := io.Pipe()

		// Output send to Pipe
		go client.AttachToContainer(docker.AttachToContainerOptions{
			Container:    cont.Name,
			OutputStream: writer,
			Logs:         true,
			Stdout:       true,
		})

		// Consume output and send to a new queue
		go func(reader io.Reader) {
			scanner := bufio.NewScanner(reader)
			ch, err := c.conn.Channel()
			if err != nil {
				ErrorOnResponseQueue(err.Error())
			}

			q, err := ch.QueueDeclare("response", true, false, false, false, nil)
			if err != nil {
				ErrorOnResponseQueue(err.Error())
			}
			for scanner.Scan() {
				ch.Publish("", q.Name, false, false, amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(scanner.Text()),
				})
			}

		}(reader)

		d.Ack(true)
	}
	log.Printf("handle: deliveries channel closed")
	c.done <- nil
}
