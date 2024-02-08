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
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Reconciler) syncDeleteTestCase(ctx context.Context, testCase *tofaniov1alpha1.TestCase) (result reconcile.Result, err error) {
	if controllerutil.ContainsFinalizer(testCase, constants.TofanObjectTemplateFinalizer) {
		controllerutil.RemoveFinalizer(testCase, constants.TofanObjectTemplateFinalizer)

		if err = r.Update(ctx, testCase); err != nil {
			return ctrl.Result{}, err

		}
	}
	return ctrl.Result{}, err

}

func (r *Reconciler) syncTestCase(ctx context.Context, testCase *tofaniov1alpha1.TestCase) (result reconcile.Result, err error) {
	if !controllerutil.ContainsFinalizer(testCase, constants.TofanObjectTemplateFinalizer) {
		controllerutil.AddFinalizer(testCase, constants.TofanObjectTemplateFinalizer)

		objectTemplate, err := r.FetchObjectTemplate(ctx, testCase.Namespace, testCase.Spec.ObjectTemplateRef.Name)
		if err != nil {
			// Handle error accordingly, maybe requeue or set an error status on the TestCase
			return ctrl.Result{}, err
		}
		err = r.ProcessTestCase(objectTemplate, testCase)
		if err != nil {
			return reconcile.Result{}, err
		}
		//r.Log.Info("modifiedTemplate", modifiedTemplate)
		if err = r.Update(ctx, testCase); err != nil {
			r.Log.Info("Reconciling TestCase")

			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{
		RequeueAfter: constants.RequeueAfter,
	}, nil

}
