package testcase

import (
	"context"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

// startReadinessWatcher initiates a goroutine that periodically checks the readiness of resources
// associated with the given TestCase and ObjectTemplate. It uses a ticker to perform checks
func (r *Reconciler) startReadinessWatcher(ctx context.Context, testCase *tofaniov1alpha1.TestCase, objTpl *tofaniov1alpha1.ObjectTemplate) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				allReady, err := r.CheckTestCaseResourcesReadiness(ctx, testCase, objTpl)
				if err != nil {
					r.Log.Error(err, "Error checking resource readiness", "TestCase", testCase.Name)
					continue
				}
				r.Log.Info("Resource readiness check result", "TestCase", testCase.Name, "AllReady", allReady)

				if allReady {
					r.EmitEvent(testCase, testCase.GetName(), controllerutil.OperationResultUpdatedStatus, StatusCompletedMsg, nil)
					r.ProcessCondition(ctx, testCase, constants.ObjConditionReady, metav1.ConditionTrue, StatusCompletedReason, StatusCompletedMsg)

					err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
						// Re-fetch the latest version of testCase before attempting update
						updatedTestCase := &tofaniov1alpha1.TestCase{}
						if err := r.Get(ctx, types.NamespacedName{Name: testCase.Name, Namespace: testCase.Namespace}, updatedTestCase); err != nil {
							return err
						}
						updatedTestCase.Status.Phase = StatusCompleted
						return r.Status().Update(ctx, updatedTestCase)
					})

					if err != nil {
						r.Log.Error(err, "Failed to update TestCase status to completed after retries", "TestCase", testCase.Name)
					} else {
						if err := r.TeardownResourcesForTestCase(ctx, testCase, objTpl); err != nil {
							r.Log.Error(err, "Failed to teardown resources", "TestCase", testCase.Name)
						}
						r.Log.Info("Readiness confirmed and teardown completed successfully", "TestCase", testCase.Name)
					}
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
