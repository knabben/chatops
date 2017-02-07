package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/pkg/api/v1"
)

// Shared default namespace
var ns string = "backing"

func createPod(clientset *kubernetes.Clientset, event Event, msg amqp.Delivery, envList []varEnv) *v1.Pod {
	var podName = event.Id

	// Create Pod Spec in v1.Pod, extracts environment variables from vars.yaml
	pod := &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
			Labels:    map[string]string{"name": podName},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:  podName,
					Image: event.Image,
					Args:  event.Args,
				},
			},
		},
	}

	envVars := []v1.EnvVar{}
	for _, env := range envList {
		envVars = append(envVars, v1.EnvVar{
			Name: env.Name,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: env.Secret,
					},
					Key: env.Key,
				},
			},
		})
	}

	if len(envVars) != 0 {
		pod.Spec.Containers[0].Env = envVars
	}

	// Run the pod and print error
	log.Println("Creating POD: ", pod)
	podObject, err := clientset.Pods(ns).Create(pod)
	if err != nil {
		msg.Ack(true)
		log.Fatal("ERROR: ", err)
	}
	return podObject
}

func ReadLogAndPublish(clientset *kubernetes.Clientset, consumer Consumer, pod *v1.Pod) {
	channel, err := consumer.conn.Channel()
	if err != nil {
		fmt.Errorf("ERROR: channel creation %s", err)
	}
	podName := pod.ObjectMeta.Name
	queue, err := channel.QueueDeclare("response", true, false, false, false, nil)
	for {
		time.Sleep(3000 * time.Millisecond)
		result, _ := clientset.Pods(ns).Get(podName)
		// TODO - Complex checking for ContainerStateWaiting (missing a Secret
		// 		  for example
		if result.Status.Phase == "Succeeded" || result.Status.Phase == "Failed" {
			body, _ := clientset.Pods(ns).GetLogs(podName, &v1.PodLogOptions{}).Do().Raw()
			channel.Publish("", queue.Name, false, false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)
			break
		}
	}
}

func deletePod(clientset *kubernetes.Clientset, podName string) {
	err := clientset.Pods(ns).Delete(podName, nil)
	fmt.Println("Deleting pod -> ", podName)
	if err != nil {
		fmt.Println("ERROR: trying to delete pod ", err)
	}
}
