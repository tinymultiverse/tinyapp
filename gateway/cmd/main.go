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
	"net/http"

	"github.com/tinymultiverse/tinyapp/gateway/internal"
	"github.com/tinymultiverse/tinyapp/gateway/proxy"
	"github.com/tinymultiverse/tinyapp/gateway/util/metrics"
	"github.com/tinymultiverse/tinyapp/util"
	"github.com/tinymultiverse/tinyapp/util/logging"

	"github.com/caarlos0/env/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var envVars internal.EnvVars

func main() {
	logging.InitLoggerFromEnvironment()

	envVars = internal.EnvVars{}
	if err := env.Parse(&envVars); err != nil {
		zap.S().Fatalw("could not process environment variables", "error", err)
	}

	prometheus.MustRegister(metrics.UsernameCounter)

	proxyConfig, err := proxy.NewProxyServerConfig(envVars)
	if err != nil {
		zap.S().Fatalw("failed to set up proxy", "error", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", proxyConfig)
	addr := ":" + envVars.HttpPort
	go func() {
		zap.S().Info("starting proxy gateway")
		if err := http.ListenAndServe(addr, mux); err != nil {
			zap.S().Fatalw("could not start proxy server", "error", err)
		}
	}()

	if envVars.MetricsEnabled {
		if envVars.MetricsPath == "" || envVars.MetricsPort == "" {
			zap.S().Fatal("METRICS_PATH and METRICS_PORT must be set if METRICS_ENABLED is true")
		}

		metricsMux := http.NewServeMux()
		metricsMux.Handle(envVars.MetricsPath, promhttp.Handler())
		metricsAddr := ":" + envVars.MetricsPort

		if envVars.MetricsTlsEnabled {
			zap.S().Info("starting https metrics server")
			if err := http.ListenAndServeTLS(metricsAddr, util.TLSCertFile, util.TLSKeyFile, metricsMux); err != nil {
				zap.S().Fatalw("could not start metrics server", "error", err)
			}
		} else {
			zap.S().Info("starting http metrics server")
			if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
				zap.S().Fatalw("could not start metrics server", "error", err)
			}
		}
	}
}
