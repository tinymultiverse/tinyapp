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

package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/tinymultiverse/tinyapp/controller/util"
	"github.com/tinymultiverse/tinyapp/gateway/internal"
	"github.com/tinymultiverse/tinyapp/gateway/util/metrics"
	globalutil "github.com/tinymultiverse/tinyapp/util"
	"go.uber.org/zap"
)

type proxyServerConfig struct {
	Proxy      *httputil.ReverseProxy
	URLSubPath string
}

func NewProxyServerConfig(envVars internal.EnvVars) (*proxyServerConfig, error) {
	targetURL, err := url.Parse("http://localhost:" + util.DefaultAppPort)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return &proxyServerConfig{
		Proxy:      proxy,
		URLSubPath: envVars.URLSubPath,
	}, nil
}

func (p *proxyServerConfig) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	zap.S().Debugw("got a request", "host", req.Host, "method", req.Method, "requestURL", req.URL.String())

	// Only increment user count if the request URL is app homepage
	if path.Clean(req.URL.Path) == path.Clean(p.URLSubPath) {
		zap.S().Info("Incrementing user count")
		// TODO Once integrated with OAuth, get actual username from auth server
		metrics.UsernameCounter.WithLabelValues(globalutil.AnyUserName).Inc()
	}

	p.Proxy.ServeHTTP(res, req)
}
