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

package main

import (
	"github.com/tinymultiverse/tinyapp/controller/internal"
	"github.com/tinymultiverse/tinyapp/controller/reconciler"
	"github.com/tinymultiverse/tinyapp/controller/util"
	v1alpha12 "github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	"github.com/tinymultiverse/tinyapp/util/logging"

	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Fix 'no Auth Provider found for name \"oidc\"'
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var envVars internal.EnvVars

func init() {
	logging.InitLoggerFromEnvironment()

	envVars = internal.EnvVars{}
	if err := env.Parse(&envVars); err != nil {
		zap.S().Fatalw("could not process environment variables", "error", err)
	}

	if envVars.IngressAnnotations == nil {
		envVars.IngressAnnotations = map[string]string{}
	}

	if envVars.PodAnnotations == nil {
		envVars.PodAnnotations = map[string]string{}
	}
}

func main() {
	zap.S().Info("Initializing TinyApp controller")

	opts := ctrl.Options{
		Namespace:              envVars.TinyAppNamespace,
		MetricsBindAddress:     ":" + envVars.ControllerMetricsPort,
		HealthProbeBindAddress: ":8082",
	}

	// Instantiate controllers manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opts)
	if err != nil {
		zap.S().Fatalw("failed to instantiate new TinyApp controller manager", "error", err)
	}

	// Register controllers
	if err = v1alpha12.AddToScheme(mgr.GetScheme()); err != nil {
		zap.S().Fatalw("failed to register TinyApp controller scheme", "error", err)
	}

	// Add a readiness endpoint
	if err = mgr.AddReadyzCheck("healthz", healthz.Ping); err != nil {
		zap.S().Fatalw("failed to add readiness check", "error", err)
	}

	// Add a health endpoint
	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		zap.S().Fatalw("failed to add health check", "error", err)
	}

	// Instantiate controllers
	k8sConfig := config.GetConfigOrDie()
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		zap.S().Fatalw("could not instantiate interface for TinyApp dependents")
	}

	c, err := controller.New(util.ControllerName, mgr, controller.Options{
		Reconciler: reconciler.NewReconciler(mgr.GetClient(), k8sClient, envVars),
	})

	if err != nil {
		zap.S().Fatalw("could not instantiate TinyApp reconciler", "error", err)
	}

	// Watch for TinyApp
	if err = c.Watch(
		&source.Kind{Type: &v1alpha12.TinyApp{}}, &handler.EnqueueRequestForObject{}); err != nil {
		zap.S().Fatalw("failed to register TinyApp watcher", "error", err)
	}

	dependentPredicate := predicate.Or(
		predicate.GenerationChangedPredicate{},
		predicate.LabelChangedPredicate{})

	// Watch for deployment
	if err = c.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha12.TinyApp{}},
		dependentPredicate); err != nil {
		zap.S().Fatalw("failed to register Deployment watcher", "error", err)
	}

	// Watch for service
	if err = c.Watch(
		&source.Kind{Type: &corev1.Service{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha12.TinyApp{}},
		dependentPredicate); err != nil {
		zap.S().Fatalw("failed to register Service watcher", "error", err)
	}

	// Watch for ingress
	if err = c.Watch(
		&source.Kind{Type: &networkingv1.Ingress{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha12.TinyApp{}},
		dependentPredicate); err != nil {
		zap.S().Fatalw("failed to register Service watcher", "error", err)
	}

	zap.S().Info("Starting TinyApp controller")
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		zap.S().Fatalw("failed to start controllers manager", "error", err)
	}
}
