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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chatv1 "github.com/api/v1"
)

// ChatReconciler reconciles a Chat object
type ChatReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=chat.chat.ops,resources=chats,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chat.chat.ops,resources=chats/status,verbs=get;update;patch

func (r *ChatReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("chat", req.NamespacedName)

	chat := &chatv1.Chat{}
	if err := r.Get(ctx, req.NamespacedName, chat); err != nil {
		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}
	fmt.Println(chat)
	return ctrl.Result{}, nil
}

func (r *ChatReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1.Chat{}).
		Complete(r)
}
