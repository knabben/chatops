package chat

import (
	"context"
	"fmt"
	chatv1alpha1 "github.com/knabben/chatops/pkg/apis/chat/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_chat")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Chat Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, outputChan chan string) error {
	return add(mgr, newReconciler(mgr, outputChan))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, outputChan chan string) reconcile.Reconciler {
	return &ReconcileChat{client: mgr.GetClient(), scheme: mgr.GetScheme(), outputChan: outputChan}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("chat-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Chat
	err = c.Watch(&source.Kind{Type: &chatv1alpha1.Chat{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileChat implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileChat{}

// ReconcileChat reconciles a Chat object
type ReconcileChat struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	outputChan chan string
}

// Reconcile reads that state of the cluster for a Chat object and makes changes based on the state read
// and what is in the Chat.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileChat) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Chat")

	// Fetch the Chat instance
	instance := &chatv1alpha1.Chat{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	podList := &corev1.PodList{}
	err = r.client.List(context.Background(), nil, podList)
	if err != nil {
		fmt.Println(err)
	}

	for _, podItem := range podList.Items {
		if strings.Contains(podItem.ObjectMeta.Name, instance.Status.Command) {
			fmt.Println(podItem)
			fmt.Println(instance.Spec.Command)
		}
	}

	return reconcile.Result{}, nil
}