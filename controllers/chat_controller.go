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
	//"bytes"
	//"github.com/knabben/chatops/pkg/chat"
	//"io"
	//"net/url"
	//"strings"

	//"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	chatv1 "github.com/knabben/chatops/api/v1"
	"github.com/knabben/chatops/pkg/chat"
	"github.com/knabben/chatops/pkg/command"

	//"github.com/knabben/chatops/pkg/chat"
	//"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"

	"github.com/spf13/viper"
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

	clientset, err := kubernetes.NewForConfig(r.Config)
	if err != nil {
		fmt.Println(err)
		return ctrl.Result{}, err
	}

	var chatType chatv1.Chat
	if err := r.Get(ctx, req.NamespacedName, &chatType); err != nil {
		log.Error(err, "unable to fetch chat type.")
		return ctrl.Result{}, nil
	}

	pod := r.FindPod(ctx, chatType.Spec.PodLabel, req)
	if pod == nil {
		log.Error(err, "unable to find pod.")
		return ctrl.Result{}, nil
	}

	podExec := command.NewPodExec(r.Config, *clientset, req.Namespace)
	stdout, _, err := podExec.ExecCommandInContainer(pod.ObjectMeta.Name, chatType.Status.Command, chatType.Status.Arguments)

	chatClient := chat.NewChat(viper.GetString("slack_token"), r.Client)
	chatClient.SendMessage(stdout)

	return ctrl.Result{}, nil
}

// FindPod search for a pod by label - how to use the lib for this
func (r *ChatReconciler) FindPod(ctx context.Context, label string, req ctrl.Request) *corev1.Pod {
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace)); err != nil {
		r.Log.Error(err, "unable to list child Jobs")
		//return ctrl.Result{}, err
	}
	for _, pod := range childPods.Items {
		for _, podLabel := range pod.ObjectMeta.Labels {
			if podLabel == label {
				return &pod
			}
		}
	}
	return nil
}


func (r *ChatReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1.Chat{}).
		Complete(r)
}
