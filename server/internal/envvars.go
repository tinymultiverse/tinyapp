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

// EnvVars are used throughout code
type EnvVars struct {
	GRPCPort              int    `env:"GRPC_PORT" envDefault:"8082"`
	HTTPPort              int    `env:"HTTP_PORT" envDefault:"8889"`
	KubeConfigPath        string `env:"KUBE_CONFIG_PATH"`
	TinyAppNamespace      string `env:"TINY_APP_NAMESPACE,notEmpty"`
	DefaultAppImage       string `env:"DEFAULT_APP_IMAGE"`
	AppIngressDomain      string `env:"APP_INGRESS_DOMAIN,notEmpty"`
	AppIngressSubPath     string `env:"APP_INGRESS_SUB_PATH"`
	AppIngressTlsEnabled  bool   `env:"APP_INGRESS_TLS_ENABLED" envDefault:"true"`
	PrometheusUserName    string `env:"PROMETHEUS_USER_NAME"` // Ignored if PrometheusSecretPath is set
	PrometheusPassword    string `env:"PROMETHEUS_PASSWORD"`  // Ignored if PrometheusSecretPath is set
	PrometheusSecretPath  string `env:"PROMETHEUS_SECRET_PATH"`
	PrometheusUrl         string `env:"PROMETHEUS_URL"`           // Required if utilizing metrics endpoints
	DefaultGitTokenSecret string `env:"DEFAULT_GIT_TOKEN_SECRET"` // Default k8s secret name for git token
}
