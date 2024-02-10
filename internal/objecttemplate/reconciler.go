package objecttemplate

import (
	"context"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/constants"
	"github.com/invioteq/tofan/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Reconciler) syncObjectTemplate(ctx context.Context, objecTpl *tofaniov1alpha1.ObjectTemplate) (result reconcile.Result, err error) {
	if !controllerutil.ContainsFinalizer(objecTpl, constants.TofanObjectTemplateFinalizer) {
		controllerutil.AddFinalizer(objecTpl, constants.TofanObjectTemplateFinalizer)

		if err = r.Update(ctx, objecTpl); err != nil {
			r.Log.Info("Reconciling ObjectTemplate")
			r.EmitEvent(objecTpl, objecTpl.GetName(), controllerutil.OperationResultCreated, "Creating ObjectTemplate in progress", nil)
			return ctrl.Result{}, err
		}
	}

	r.ProcessCondition(ctx, objecTpl, constants.ObjConditionReady, metav1.ConditionTrue, "ObjectTemplateSyncSuccess", "ObjectTemplate synced successfully")
	// Update the ObjectTpl status with kind & Group
	ObjKind, ObjApiVersion, err := utils.ExtractKindAndAPIVersion(objecTpl)

	objecTpl.Status.Group = ObjApiVersion
	objecTpl.Status.Kind = ObjKind
	err = r.UpdateStatus(ctx, objecTpl)
	if err != nil {
		r.Log.Info("error updating the status")
	}
	r.EmitEvent(objecTpl, objecTpl.GetName(), controllerutil.OperationResultUpdatedStatus, "ObjectTemplate synced successfully", nil)

	return ctrl.Result{
		RequeueAfter: constants.RequeueAfter,
	}, nil

}

func (r *Reconciler) syncDeleteObjectTemplate(ctx context.Context, objecTpl *tofaniov1alpha1.ObjectTemplate) (result reconcile.Result, err error) {
	if controllerutil.ContainsFinalizer(objecTpl, constants.TofanObjectTemplateFinalizer) {
		controllerutil.RemoveFinalizer(objecTpl, constants.TofanObjectTemplateFinalizer)

		if err = r.Update(ctx, objecTpl); err != nil {
			return ctrl.Result{}, err

		}
	}
	return ctrl.Result{}, err

}
