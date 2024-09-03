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

package internal

// EnvVars used throughout code
type EnvVars struct {
	TinyAppNamespace  string            `env:"TINY_APP_NAMESPACE,notEmpty""`
	AppServiceAccount string            `env:"APP_SERVICE_ACCOUNT" envDefault:"default"`
	GitSyncImage      string            `env:"GIT_SYNC_IMAGE" envDefault:"registry.k8s.io/git-sync/git-sync:v3.6.8"`
	GitSyncEnvVars    map[string]string `env:"GIT_SYNC_ENV_VARS" envKeyValSeparator:"="`
	GatewayEnvVars    map[string]string `env:"GATEWAY_ENV_VARS" envKeyValSeparator:"="`
	GatewayImage      string            `env:"GATEWAY_IMAGE" envDefault:"quay.io/tinymultiverse/tinyapp-gateway:latest"`
	// Required if GATEWAY_METRICS_TLS_ENABLED is true or any app ingress will have TLS enabled
	TLSSecretName            string            `env:"TLS_SECRET_NAME"`
	GatewayMetricsEnabled    bool              `env:"GATEWAY_METRICS_ENABLED" envDefault:"true"`
	GatewayMetricsTlsEnabled bool              `env:"GATEWAY_METRICS_TLS_ENABLED" envDefault:"false"`
	GatewayMetricsPort       string            `env:"GATEWAY_METRICS_PORT" envDefault:"9090"`
	GatewayMetricsPath       string            `env:"GATEWAY_METRICS_PATH" envDefault:"/metrics"`
	ControllerMetricsPort    string            `env:"CONTROLLER_METRICS_PORT" envDefault:"8085"`
	DefaultAppEnvVars        map[string]string `env:"DEFAULT_APP_ENV_VARS" envKeyValSeparator:"="`
	PodAnnotations           map[string]string `env:"POD_ANNOTATIONS" envKeyValSeparator:"="`
	IngressAnnotations       map[string]string `env:"INGRESS_ANNOTATIONS" envKeyValSeparator:"="`
}
