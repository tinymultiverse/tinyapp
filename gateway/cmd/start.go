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

package cmd

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

func Start() {
	logging.InitLoggerFromEnvironment()

	envVars = internal.EnvVars{}
	if err := env.Parse(&envVars); err != nil {
		zap.S().Fatalw("could not process environment variables", "error", err)
	}

	// Validate OIDC configuration if enabled
	if envVars.OIDCEnabled {
		if envVars.OIDCIssuerURL == "" || envVars.OIDCClientID == "" ||
			envVars.OIDCClientSecret == "" || envVars.OIDCRedirectURL == "" {
			zap.S().Fatal("OIDC_ISSUER_URL, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL must be set if OIDC_ENABLED is true")
		}
	}

	proxyConfig, err := proxy.NewProxyServerConfig(envVars)
	if err != nil {
		zap.S().Fatalw("failed to set up proxy", "error", err)
	}

	mux := http.NewServeMux()

	// Apply OIDC middleware if enabled
	var handler http.Handler = proxyConfig
	if proxyConfig.OIDCAuth != nil {
		if envVars.AuthzEnabled {
			zap.S().Info("OIDC authentication and authorization enabled")
			handler = proxyConfig.OIDCAuth.AuthorizationMiddleware(proxyConfig)
		} else {
			zap.S().Info("OIDC authentication enabled (no authorization)")
			handler = proxyConfig.OIDCAuth.Middleware(proxyConfig)
		}
	}

	mux.Handle("/", handler)
	addr := ":" + envVars.HttpPort

	if envVars.MetricsEnabled {
		if envVars.MetricsPath == "" || envVars.MetricsPort == "" {
			zap.S().Fatal("METRICS_PATH and METRICS_PORT must be set if METRICS_ENABLED is true")
		}

		prometheus.MustRegister(metrics.UsernameCounter)

		metricsMux := http.NewServeMux()
		metricsMux.Handle(envVars.MetricsPath, promhttp.Handler())
		metricsAddr := ":" + envVars.MetricsPort

		go func() {
			// Start the proxy server in a separate goroutine
			startProxyServer(addr, mux)
		}()

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
	} else {
		// Start the proxy server directly if metrics are not enabled
		startProxyServer(addr, mux)
	}
}

func startProxyServer(addr string, mux *http.ServeMux) {
	zap.S().Infow("starting proxy gateway", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		zap.S().Fatalw("could not start proxy server", "error", err)
	}
}
