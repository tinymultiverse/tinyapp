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
	"strings"

	"github.com/tinymultiverse/tinyapp/gateway/auth"
	"github.com/tinymultiverse/tinyapp/gateway/internal"
	"github.com/tinymultiverse/tinyapp/gateway/util/metrics"
	globalutil "github.com/tinymultiverse/tinyapp/util"
	"go.uber.org/zap"
)

type proxyServerConfig struct {
	Proxy                  *httputil.ReverseProxy
	SecondaryProxy         *httputil.ReverseProxy
	SecondaryTargetPattern string
	URLSubPath             string
	authenticator          *auth.LDAPAuthenticator
}

func NewProxyServerConfig(envVars internal.EnvVars) (*proxyServerConfig, error) {
	targetURL, err := url.Parse("http://127.0.0.1:" + envVars.PrimaryTargetPort)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	var secondaryProxy *httputil.ReverseProxy
	if envVars.SecondaryTargetPattern != "" && envVars.SecondaryTargetPort != "" {
		secondaryTargetUrl := &url.URL{
			Scheme: "http",
			Host:   "127.0.0.1:" + envVars.SecondaryTargetPort,
		}
		secondaryProxy = httputil.NewSingleHostReverseProxy(secondaryTargetUrl)
	}

	authenticator := auth.NewLDAPAuthenticator(envVars)

	return &proxyServerConfig{
		Proxy:                  proxy,
		SecondaryProxy:         secondaryProxy,
		SecondaryTargetPattern: envVars.SecondaryTargetPattern,
		URLSubPath:             envVars.URLSubPath,
		authenticator:          authenticator,
	}, nil
}

func (p *proxyServerConfig) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	zap.S().Debugw("got a request", "host", req.Host, "method", req.Method, "requestURL", req.URL.String())

	username, err := p.authenticator.Authenticate(req)
	if err != nil {
		zap.S().Warnw("authentication failed", "username", username, "error", err, "remoteAddr", req.RemoteAddr)
		p.authenticator.RequireAuth(res)
		return
	}

	// Set the authenticated username for metrics and logging
	authenticatedUser := globalutil.AnyUserName
	if username != "" {
		authenticatedUser = username
		zap.S().Debugw("authenticated user", "username", username)
	}

	err = p.authenticator.AuthorizeUser(username)
	if err != nil {
		zap.S().Warnw("authorization failed", "username", username, "error", err, "remoteAddr", req.RemoteAddr)
		p.sendUnauthorizedResponse(res, username)
		return
	}

	if p.SecondaryProxy != nil && strings.Contains(req.URL.Path, p.SecondaryTargetPattern) {
		zap.S().Debugw("routing to secondary proxy", "path", req.URL.Path, "user", authenticatedUser)
		p.SecondaryProxy.ServeHTTP(res, req)
		return
	}

	// Only increment user count if the request URL is app homepage
	if path.Clean(req.URL.Path) == path.Clean(p.URLSubPath) {
		zap.S().Infow("Incrementing user count", "user", authenticatedUser)
		metrics.UsernameCounter.WithLabelValues(authenticatedUser).Inc()
	}

	p.Proxy.ServeHTTP(res, req)
}

// sendUnauthorizedResponse sends an HTML response for unauthorized users
func (p *proxyServerConfig) sendUnauthorizedResponse(res http.ResponseWriter, username string) {
	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(http.StatusForbidden)
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Access Denied</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 50px; text-align: center; }
        .container { max-width: 500px; margin: 0 auto; }
        h1 { color: #d32f2f; }
        p { color: #666; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Access Denied</h1>
        <p>You do not have access to this app.</p>
        <p>User: ` + username + `</p>
        <p>Please contact your administrator if you believe this is an error.</p>
    </div>
</body>
</html>`
	res.Write([]byte(html))
}
