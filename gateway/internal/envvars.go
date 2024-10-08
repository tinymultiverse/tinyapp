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

type EnvVars struct {
	HttpPort          string `env:"HTTP_PORT" envDefault:"8889"`
	MetricsEnabled    bool   `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsTlsEnabled bool   `env:"METRICS_TLS_ENABLED" envDefault:"false"`
	MetricsPort       string `env:"METRICS_PORT"` // Required if METRICS_ENABLED is true
	MetricsPath       string `env:"METRICS_PATH"` // Required if METRICS_ENABLED is true
	TinyAppName       string `env:"TINY_APP_NAME,notEmpty"`
}
