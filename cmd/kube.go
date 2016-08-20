package cmd

import (
	"fmt"
	"github.com/streadway/amqp"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"time"
)

var ns = api.NamespaceDefault

func createPod(kubeClient *client.Client, event Event, podName string) {
	// Pod specification
	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:      podName,
			Namespace: ns,
			Labels:    map[string]string{"name": podName},
		},
		Spec: api.PodSpec{
			RestartPolicy: api.RestartPolicyOnFailure,
			Containers: []api.Container{
				{
					Name:  event.Pod,
					Image: event.Image,
					Args:  event.Args,
				},
			},
		},
	}

	// Run the pod and print error
	_, err := kubeClient.Pods(ns).Create(pod)
	if err != nil {
		fmt.Println("ERROR: [%s] with %s", podName, err)
	}
}

func ReadLogAndPublish(kubeClient *client.Client, consumer Consumer, podName string) {
	channel, err := consumer.conn.Channel()
	if err != nil {
		fmt.Errorf("ERROR: channel creation %s", err)
	}

	queue, err := channel.QueueDeclare("response", true, false, false, false, nil)
	for {
		body, _ := kubeClient.Pods(api.NamespaceDefault).GetLogs(podName, &api.PodLogOptions{}).Do().Raw()
		// TODO - First log output, check pod state before ending
		if len(body) > 0 {
			channel.Publish("", queue.Name, false, false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func deletePod(kubeClient *client.Client, podName string) {
	err := kubeClient.Pods(ns).Delete(podName, nil)
	if err != nil {
		fmt.Println("ERROR: trying to delete pod ", err)
	}
}
