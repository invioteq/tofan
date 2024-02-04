/*
Copyright 2024 invioteq llc.

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

package testcase

import (
	"context"
	"github.com/invioteq/tofan/internal/common"

	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconciler  reconciles a TestCase object
type Reconciler struct {
	common.Reconciler
}

//+kubebuilder:rbac:groups=tofan.io,resources=testcases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tofan.io,resources=testcases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tofan.io,resources=testcases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("TestCase", req.NamespacedName)

	// Fetch the TestCase resource
	testCase := &tofaniov1alpha1.TestCase{}

	err := r.Get(ctx, req.NamespacedName, testCase)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// TestCase not found, return
			log.Info("TestCase not found.")

			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	if !testCase.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.syncTestCase(ctx, testCase)
	}
	return r.syncTestCase(ctx, testCase)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tofaniov1alpha1.TestCase{}).
		Complete(r)
}
