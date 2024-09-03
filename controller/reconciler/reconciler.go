/*
Copyright 2024 BlackRock, Inc.

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

package reconciler

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tinymultiverse/tinyapp/controller/reconciler/builder"
	"github.com/tinymultiverse/tinyapp/controller/util"
	"github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tinymultiverse/tinyapp/controller/internal"
	"go.uber.org/zap"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler struct {
	tinyAppClient client.Client
	k8sClient     kubernetes.Interface
	env           internal.EnvVars
}

// NewReconciler returns a reconciler
func NewReconciler(tinyAppClient client.Client, k8sClient kubernetes.Interface, env internal.EnvVars) *reconciler {
	workqueue.DefaultControllerRateLimiter()
	return &reconciler{
		tinyAppClient: tinyAppClient,
		k8sClient:     k8sClient,
		env:           env,
	}
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := zap.S().With("app", request.NamespacedName)

	tinyApp := &v1alpha1.TinyApp{}

	logger.Debugw("Received a reconcile request for TinyApp")

	// Make sure object exists before we operate on it
	err := r.tinyAppClient.Get(ctx, request.NamespacedName, tinyApp)
	if err != nil {
		if k8sErrors.IsNotFound(err) { // Assume object was just deleted
			logger.Info("TinyApp not found for reconciliation, so likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Errorf("Unable to get TinyApp for reconciliation: %v", err)
		return reconcile.Result{}, err
	}

	logger.Debug("Retrieved TinyApp")

	if tinyApp.DeletionTimestamp != nil {
		logger.Debugw("ForegroundDeletion is in process", "deletionTimestamp", tinyApp.DeletionTimestamp)
		return reconcile.Result{}, nil
	}

	// Reconcile TinyApp state
	err = r.reconcileTinyAppState(ctx, tinyApp)
	if err != nil {
		logger.Errorw("Failed to reconcile TinyApp state", "name", tinyApp.GetName(), "error", err)
	}

	logger.Infow("Reconcile request completed", "app status", tinyApp.Status)

	return reconcile.Result{}, err
}

// reconcileTinyAppState updates TinyApp status and all its dependents.
func (r *reconciler) reconcileTinyAppState(ctx context.Context, app *v1alpha1.TinyApp) error {
	logger := zap.S().With("app", app.Name)

	defer r.updateAppStatus(ctx, app)
	app.Status.InitConditions()

	logger.Debug("Reconciling service")
	if err := r.reconcileService(ctx, app); err != nil {
		app.Status.SetConditionFalseWithMessage(v1alpha1.ServiceCreated, err.Error())
		return err
	}
	app.Status.SetConditionTrue(v1alpha1.ServiceCreated)

	logger.Debug("Reconciling ingress")
	if err := r.reconcileIngress(ctx, app); err != nil {
		app.Status.SetConditionFalseWithMessage(v1alpha1.IngressCreated, err.Error())
		return err
	}
	app.Status.SetConditionTrue(v1alpha1.IngressCreated)

	logger.Debug("Reconciling deployment")
	if err := r.reconcileDeployment(ctx, app); err != nil {
		app.Status.SetConditionFalseWithMessage(v1alpha1.DeploymentCreated, err.Error())
		return err
	}
	app.Status.SetConditionTrue(v1alpha1.DeploymentCreated)

	return nil
}

func (r *reconciler) updateAppStatus(ctx context.Context, app *v1alpha1.TinyApp) {
	app.Status.Phase = app.Status.GetPhase()

	err := r.tinyAppClient.Status().Update(ctx, app)
	if err != nil {
		zap.S().Errorw("Failed to update TinyApp status", "name", app.Name, "error", err)
		return
	}

	zap.S().Infow("TinyApp status updated", "name", app.Name)
}

func (r *reconciler) reconcileService(ctx context.Context, app *v1alpha1.TinyApp) error {
	service, err := r.k8sClient.CoreV1().Services(r.env.TinyAppNamespace).Get(ctx, app.Name, metav1.GetOptions{})
	if err != nil {
		// Service does not exist - create it
		if k8sErrors.IsNotFound(err) {
			err = r.createService(&ctx, app)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Service exists - first check if update is needed
		if !r.shouldPerformServiceUpdate(app, service) {
			zap.S().Infow("Service update not needed", "name", app.Name)
			return nil
		}

		// Update service
		err = r.updateService(&ctx, app)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) createService(ctx *context.Context, app *v1alpha1.TinyApp) error {
	zap.S().Infow("Creating service for TinyApp", "name", app.Name)

	service, err := builder.BuildService(app, r.env)
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp service object")
	}

	_, err = r.k8sClient.CoreV1().Services(r.env.TinyAppNamespace).Create(*ctx, service, metav1.CreateOptions{})
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp service")
	}

	return nil
}

func (r *reconciler) updateService(ctx *context.Context, app *v1alpha1.TinyApp) error {
	zap.S().Infow("Updating service for TinyApp", "name", app.Name)

	newService, err := builder.BuildService(app, r.env)
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp service object")
	}

	_, err = r.k8sClient.CoreV1().Services(r.env.TinyAppNamespace).Update(*ctx, newService, metav1.UpdateOptions{})
	if err != nil {
		return errors.WithMessage(err, "failed to update TinyApp service")
	}

	return nil
}

func (r *reconciler) reconcileIngress(ctx context.Context, app *v1alpha1.TinyApp) error {
	ingress, err := r.k8sClient.NetworkingV1().Ingresses(r.env.TinyAppNamespace).Get(ctx, app.Name, metav1.GetOptions{})
	if err != nil {
		// Ingress does not exist - create it
		if k8sErrors.IsNotFound(err) {
			err = r.createIngress(&ctx, app)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Ingress exists - first check if update is needed
		if !r.shouldPerformIngressUpdate(app, ingress) {
			zap.S().Infow("Ingress update not needed", "name", app.Name)
			return nil
		}

		// Update ingress
		err = r.updateIngress(&ctx, app)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) createIngress(ctx *context.Context, app *v1alpha1.TinyApp) error {
	zap.S().Infow("Creating ingress for TinyApp", "name", app.Name)

	ingress, err := builder.BuildIngress(app, r.env)
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp ingress object")
	}

	_, err = r.k8sClient.NetworkingV1().Ingresses(r.env.TinyAppNamespace).Create(*ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp ingress")
	}

	return nil
}

func (r *reconciler) updateIngress(ctx *context.Context, app *v1alpha1.TinyApp) error {
	zap.S().Infow("Updating ingress for TinyApp", "name", app.Name)

	newIngress, err := builder.BuildIngress(app, r.env)
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp ingress object")
	}

	_, err = r.k8sClient.NetworkingV1().Ingresses(r.env.TinyAppNamespace).Update(*ctx, newIngress, metav1.UpdateOptions{})
	if err != nil {
		return errors.WithMessage(err, "failed to update TinyApp ingress")
	}

	return nil
}

func (r *reconciler) reconcileDeployment(ctx context.Context, app *v1alpha1.TinyApp) error {
	deployment, err := r.k8sClient.AppsV1().Deployments(r.env.TinyAppNamespace).Get(ctx, app.Name, metav1.GetOptions{})
	if err != nil {
		// Deployment does not exist - create it
		if k8sErrors.IsNotFound(err) {
			err = r.createDeployment(&ctx, app)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Deployment exists - first check if update is needed
		if !r.shouldPerformDeploymentUpdate(app, deployment) {
			zap.S().Infow("Deployment update not needed", "name", app.Name)
			return nil
		}

		// Update deployment
		err = r.updateDeployment(&ctx, app)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) createDeployment(ctx *context.Context, app *v1alpha1.TinyApp) error {
	zap.S().Infow("Creating deployment for TinyApp", "name", app.Name)

	tinyAppDeployment, err := builder.BuildDeployment(app, r.env)
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp deployment object")
	}

	_, err = r.k8sClient.AppsV1().Deployments(r.env.TinyAppNamespace).Create(*ctx, tinyAppDeployment, metav1.CreateOptions{})
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp deployment")
	}

	return nil
}

func (r *reconciler) updateDeployment(ctx *context.Context, app *v1alpha1.TinyApp) error {
	zap.S().Infow("Updating deployment for TinyApp", "name", app.Name)

	newDeployment, err := builder.BuildDeployment(app, r.env)
	if err != nil {
		return errors.WithMessage(err, "failed to create TinyApp deployment object")
	}

	_, err = r.k8sClient.AppsV1().Deployments(r.env.TinyAppNamespace).Update(*ctx, newDeployment, metav1.UpdateOptions{})
	if err != nil {
		return errors.WithMessage(err, "failed to update TinyApp deployment")
	}

	return nil
}

func (r *reconciler) shouldPerformDeploymentUpdate(app *v1alpha1.TinyApp, currentAppDeployment *appsv1.Deployment) bool {
	desiredTinyAppDeployment, _ := builder.BuildDeployment(app, r.env)
	if desiredTinyAppDeployment.Annotations[util.AnnotationResourceHash] != currentAppDeployment.Annotations[util.AnnotationResourceHash] {
		zap.S().Infow("Deployment hash diff detected, so will perform deployment update", "app name", app.Name)
		return true
	}
	return false
}

func (r *reconciler) shouldPerformServiceUpdate(app *v1alpha1.TinyApp, currentAppService *corev1.Service) bool {
	desiredAppService, _ := builder.BuildService(app, r.env)
	if desiredAppService.Annotations[util.AnnotationResourceHash] != currentAppService.Annotations[util.AnnotationResourceHash] {
		zap.S().Infow("Service hash diff detected, so will perform service update", "app name", app.Name)
		return true
	}
	return false
}

func (r *reconciler) shouldPerformIngressUpdate(app *v1alpha1.TinyApp, currentAppIngress *networkingv1.Ingress) bool {
	desiredAppIngress, _ := builder.BuildIngress(app, r.env)
	if desiredAppIngress.Annotations[util.AnnotationResourceHash] != currentAppIngress.Annotations[util.AnnotationResourceHash] {
		zap.S().Infow("Ingress hash diff detected, so will perform ingress update", "app name", app.Name)
		return true
	}
	return false
}
