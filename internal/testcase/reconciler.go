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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func (r *Reconciler) syncDeleteTestCase(ctx context.Context, testCase *tofaniov1alpha1.TestCase) (result reconcile.Result, err error) {
	// todo: double for Pending(stuck), Error status.phase >> trigger treadown process
	if !(testCase.Status.Phase == StatusPending || testCase.Status.Phase == StatusError) {
		objectTemplate, err := r.FetchObjectTemplate(ctx, testCase.Namespace, testCase.Spec.ObjectTemplateRef.Name)
		if err != nil {
			time.Sleep(30 * time.Second)
			if err := r.TeardownResourcesForTestCase(ctx, testCase, objectTemplate); err != nil {
				r.Log.Error(err, "Failed to teardown resources", "TestCase", testCase.Name)
			}
			r.Log.Info("Teardown completed successfully", "TestCase", testCase.Name)
		}
	}

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

		if testCase.Status.Phase == "" {
			r.EmitEvent(testCase, testCase.GetName(), controllerutil.OperationResultUpdatedStatusOnly, StatusPendingMsg, nil)
			r.ProcessCondition(ctx, testCase, constants.ObjConditionCreating, metav1.ConditionFalse, StatusPendingReason, StatusPendingMsg)
			testCase.Status.Phase = StatusPending
			err = r.UpdateStatus(ctx, testCase)
			if err != nil {
				r.Log.Info("error updating the status")
			}
		}

		objectTemplate, err := r.FetchObjectTemplate(ctx, testCase.Namespace, testCase.Spec.ObjectTemplateRef.Name)
		if err != nil {
			r.EmitEvent(testCase, testCase.GetName(), controllerutil.OperationResultUpdatedStatusOnly, "Cannot Find ObjectTemplateRef", err)
			r.ProcessCondition(ctx, testCase, constants.ObjConditionFailed, metav1.ConditionFalse, StatusPendingReason, StatusPendingMsg)
			testCase.Status.Phase = StatusPending
			err = r.UpdateStatus(ctx, testCase)
			if err != nil {
				r.Log.Info("error updating the status")
			}
			return ctrl.Result{}, err
		}

		if !(testCase.Status.Phase == StatusInProgress || testCase.Status.Phase == StatusCompleted || testCase.Status.Phase == StatusError) {
			r.EmitEvent(testCase, testCase.GetName(), controllerutil.OperationResultUpdatedStatus, StatusInProgressMsg, nil)
			r.ProcessCondition(ctx, testCase, constants.ObjConditionCreating, metav1.ConditionUnknown, StatusInProgressReason, StatusInProgressMsg)
			testCase.Status.Phase = StatusInProgress
			err = r.UpdateStatus(ctx, testCase)
			if err != nil {
				r.Log.Info("error updating the status")
			}

			err = r.ProcessTestCase(ctx, objectTemplate, testCase)
			if err == nil {
				// todo : TestCase should watch for external resources if there are all ready
				// todo: so we can mark the TestCase completed
				r.EmitEvent(testCase, testCase.GetName(), controllerutil.OperationResultUpdatedStatus, StatusCompletedMsg, nil)
				r.ProcessCondition(ctx, testCase, constants.ObjConditionReady, metav1.ConditionTrue, StatusCompletedReason, StatusCompletedMsg)
				testCase.Status.Phase = StatusCompleted
				err = r.UpdateStatus(ctx, testCase)
				if err != nil {
					r.Log.Info("error updating the status")
				}
				// todo: as soon as the testcase.spec.status.phase is marked Completed so
				// todo: then trigger TeardownResources process
				// todo: add annotation tofan.io/testcase-ttl : 5 minutes
				// todo: add annotation that allows end users to keep the Resources create by an X testCase
				// todo: tofan.io/testcase-keep-resources: (default is false)
				time.Sleep(30 * time.Second)
				if err := r.TeardownResourcesForTestCase(ctx, testCase, objectTemplate); err != nil {
					r.Log.Error(err, "Failed to teardown resources", "TestCase", testCase.Name)
				}
				r.Log.Info("Teardown completed successfully", "TestCase", testCase.Name)
				return reconcile.Result{}, err
			} else {
				r.EmitEvent(testCase, testCase.GetName(), controllerutil.OperationResultUpdatedStatus, StatusErrorMsg, err)
				r.ProcessCondition(ctx, testCase, constants.ObjConditionReady, metav1.ConditionFalse, StatusErrorReason, StatusErrorMsg)
				testCase.Status.Phase = StatusError
				r.Log.Error(err, "Failed to teardown resources for", "TestCase", testCase.Name)
				err = r.UpdateStatus(ctx, testCase)
				if err != nil {
					r.Log.Info("error updating the status")
				}
			}
		}

		if err = r.Update(ctx, testCase); err != nil {
			r.Log.Info("Reconciling TestCase")

			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{
		RequeueAfter: constants.RequeueAfter,
	}, nil

}
