/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	chatv1 "github.com/knabben/chatops/api/v1"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var (
	clientset *kubernetes.Clientset
	lock      sync.Mutex
)

// ChatReconciler reconciles a Chat object
type ChatReconciler struct {
	client.Client
	Log    logr.Logger
	Config *rest.Config
}

// +kubebuilder:rbac:groups=chat.ops.com,resources=chats,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chat.ops.com,resources=chats/status,verbs=get;update;patch

func (r *ChatReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("chat", req.NamespacedName)
	log.Info("Starting reconcile.")

	// List Pods
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, &client.ListOptions{Namespace: "default"}); err != nil {
		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	clientset, err := kubernetes.NewForConfig(r.Config)
	if err != nil {
		fmt.Println(err)
		return ctrl.Result{}, err
	}

	for _, pod := range childPods.Items {
		podLogOpts := corev1.PodLogOptions{}

		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
		podLogs, err := req.Stream()
		if err != nil {
			fmt.Println(err)
			return ctrl.Result{}, err
		}
		defer podLogs.Close()

		// Copy buf
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			fmt.Println(err)
			return ctrl.Result{}, err
		}

		//chatClient := chat.NewChat(viper.GetString("slack_token"))
		//chatClient.SendMessage(buf.String())
	}

	return ctrl.Result{}, nil
}

func (r *ChatReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1.Chat{}).
		Complete(r)
}
