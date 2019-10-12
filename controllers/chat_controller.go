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
	//"github.com/knabben/chatops/pkg/chat"
	//"github.com/spf13/viper"

	"bytes"
	"github.com/knabben/chatops/pkg/chat"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	chatv1 "github.com/knabben/chatops/api/v1"
	"k8s.io/client-go/kubernetes"


	corev1 "k8s.io/api/core/v1"
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

	chatClient := chat.NewChat(viper.GetString("slack_token"), r.Client)

	var chatType chatv1.Chat
	if err := r.Get(ctx, req.NamespacedName, &chatType); err != nil {
		log.Error(err, "unable to fetch chat type.")
		return ctrl.Result{}, nil
	}

	clientset, err := kubernetes.NewForConfig(r.Config)
	if err != nil {
		fmt.Println(err)
		return ctrl.Result{}, err
	}

	pod := r.GeneratePOD(&chatType)

	// ...and create it on the cluster
	if err := r.Create(ctx, pod); err != nil {
		log.Error(err, "unable to create pod", "job", pod)
		return ctrl.Result{}, err
	}
	// TODO - Listen for pod events
	time.Sleep(5 * time.Second)

	logs := r.PodLog(clientset, pod)
	chatClient.SendMessage(logs)
	log.V(1).Info("created pod run:", pod)

	return ctrl.Result{}, nil
}

func (r *ChatReconciler) GeneratePOD(chatType *chatv1.Chat) *corev1.Pod {
	name := fmt.Sprintf("%s-%d", chatType.Name, time.Now().Unix())
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   chatType.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy:  corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    name,
					Image:   chatType.Spec.JobImage,
					Command: []string{chatType.Spec.Command},
				},
			},
		},
	}
}

func (r *ChatReconciler) PodLog(clientset *kubernetes.Clientset, pod *corev1.Pod) string {
	podLogOpts := corev1.PodLogOptions{}

	podName := pod.ObjectMeta.Name
	podNamespace := pod.ObjectMeta.Namespace

	req1 := clientset.CoreV1().Pods(podNamespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req1.Stream()
	if err != nil {
		return ""
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return ""
	}
	return buf.String()
}

func (r *ChatReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1.Chat{}).
		Complete(r)
}
