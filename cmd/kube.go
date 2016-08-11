package cmd

import (
	"fmt"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func createJob(kubeClient *client.Client) {
	s, err := kubeClient.Services(api.NamespaceDefault).Get("hubot")
    if err != nil {
		fmt.Println("Can't get service:", err)
    }
    fmt.Println("Name:", s.Name)
    for p, _ := range s.Spec.Ports {
		fmt.Println("Port:", s.Spec.Ports[p].Port)
		fmt.Println("NodePort:", s.Spec.Ports[p].NodePort)
    }
}


