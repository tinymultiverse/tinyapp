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

package v1

import (
	"os"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/tinymultiverse/tinyapp/pkg/k8s/client/tinyapp/clientset/versioned"
	"github.com/tinymultiverse/tinyapp/server/internal"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Server struct {
	tinyAppClient versioned.Interface
	k8sClient     kubernetes.Interface
	promSecret    prometheusSecret
	env           internal.EnvVars
}

func NewServer(env internal.EnvVars, k8sConfig *rest.Config) (*Server, error) {
	tinyAppClient, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to make Kubernetes interface for TinyApp")
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create k8s client")
	}

	var promSecret prometheusSecret
	if env.PrometheusSecretPath != "" {
		promSecret, err = readPrometheusSecret(env.PrometheusSecretPath)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to read Prometheus secret")
		}
	} else {
		promSecret = prometheusSecret{
			Username: env.PrometheusUserName,
			Password: env.PrometheusPassword,
		}
	}

	return &Server{
		tinyAppClient: tinyAppClient,
		k8sClient:     k8sClient,
		promSecret:    promSecret,
		env:           env,
	}, nil
}

func readPrometheusSecret(filepath string) (prometheusSecret, error) {
	promSecret := prometheusSecret{}
	file, err := os.ReadFile(filepath)
	if err != nil {
		return promSecret, err
	}
	err = yaml.Unmarshal(file, &promSecret)
	if err != nil {
		return promSecret, err
	}
	return promSecret, nil
}
