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

	kbatch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"bytes"
	"context"
	"fmt"
	"time"

	//"fmt"
	"github.com/go-logr/logr"
	chatv1 "github.com/knabben/chatops/api/v1"
	//"github.com/knabben/chatops/pkg/chat"

	//"github.com/knabben/chatops/pkg/chat"
	//"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	//"github.com/spf13/viper"
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

	var chat chatv1.Chat
	if err := r.Get(ctx, req.NamespacedName, &chat); err != nil {
		log.Error(err, "unable to fetch chat type.")
		return ctrl.Result{}, nil
	}

	name := fmt.Sprintf("%s-%d", chat.Name, time.Now().Unix())
	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   chat.Namespace,
		},
		Spec: kbatch.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:  corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Image:   chat.Spec.JobImage,
							Command: []string{chat.Spec.Command},
							Name:    name,

						},
					},
				},
			},
		},
	}

	// ...and create it on the cluster
	if err := r.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job", "job", job)
		return ctrl.Result{}, err
	}

	log.V(1).Info("created Job for CronJob run", "job", job)
	return ctrl.Result{}, nil
}

func (r *ChatReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1.Chat{}).
		Complete(r)
}
