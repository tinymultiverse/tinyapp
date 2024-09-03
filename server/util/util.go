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

package util

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tinymultiverse/tinyapp/controller/reconciler/builder"
	"github.com/tinymultiverse/tinyapp/util"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetURLForTinyApp(domain, subPath, appId string, tlsEnabled bool) (string, error) {
	ingressPath, err := builder.BuildIngressPath(subPath, appId)
	if err != nil {
		return "", errors.WithMessage(err, "failed to build ingress path")
	}

	appUrl, err := url.JoinPath(domain, ingressPath)
	if err != nil {
		return "", errors.WithMessage(err, "failed to join domain and ingress path")
	}

	if tlsEnabled {
		appUrl = fmt.Sprintf("%s%s", util.Https, appUrl)
	} else {
		appUrl = fmt.Sprintf("%s%s", util.Http, appUrl)
	}

	return appUrl, nil
}

func GetKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
