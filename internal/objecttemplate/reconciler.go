package objecttemplate

import (
	"context"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Reconciler) syncObjectTemplate(ctx context.Context, objecTpl *tofaniov1alpha1.ObjectTemplate) (result reconcile.Result, err error) {
	if !controllerutil.ContainsFinalizer(objecTpl, constants.TofanObjectTemplateFinalizer) {
		controllerutil.AddFinalizer(objecTpl, constants.TofanObjectTemplateFinalizer)

		if err = r.Update(ctx, objecTpl); err != nil {
			r.Log.Info("Reconciling ObjectTemplate")

			return ctrl.Result{}, err
		}
	}
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
