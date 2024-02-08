package common

import (
	"context"
	"github.com/go-logr/logr"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles an object.
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
}

// FetchObjectTemplate retrieves an ObjectTemplate by name and namespace.
func (r *Reconciler) FetchObjectTemplate(ctx context.Context, namespace, name string) (*tofaniov1alpha1.ObjectTemplate, error) {
	objectTemplate := &tofaniov1alpha1.ObjectTemplate{}
	namespacedName := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	if err := r.Get(ctx, namespacedName, objectTemplate); err != nil {
		r.Log.Info("Failed to get ObjectTemplate", "Namespace", namespace, "Name", name)
		return nil, err
	}

	return objectTemplate, nil
}
