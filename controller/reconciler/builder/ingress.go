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

package builder

import (
	"net/url"
	"strings"

	"github.com/tinymultiverse/tinyapp/controller/internal"
	"github.com/tinymultiverse/tinyapp/controller/util"
	"github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"

	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildIngress(app *v1alpha1.TinyApp, env internal.EnvVars) (*networkingv1.Ingress, error) {
	pathType := networkingv1.PathTypeImplementationSpecific

	ingressPath, err := BuildIngressPath(app.Spec.IngressSubPath, app.Name)
	if err != nil {
		return nil, err
	}
	ingressPath = ingressPath + "(/|$)(.*)"

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:     env.IngressAnnotations,
			Name:            app.Name,
			Namespace:       env.TinyAppNamespace,
			Labels:          app.GetLabels(),
			OwnerReferences: createOwnerRefs(app),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: app.Spec.IngressDomain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: ingressPath,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: app.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: util.DefaultGatewayPort,
											},
										},
									},
									PathType: &pathType,
								},
							},
						},
					},
				},
			},
		},
	}

	if app.Spec.IngressTlsEnabled {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{app.Spec.IngressDomain},
				SecretName: env.TLSSecretName,
			},
		}
	}

	hash, err := getObjectHash(ingress)
	if err != nil {
		return nil, err
	}
	ingress.Annotations[util.AnnotationResourceHash] = hash

	return ingress, nil
}

func BuildIngressPath(subPath, appId string) (string, error) {
	if strings.TrimSpace(subPath) == "" {
		return url.JoinPath("/", appId)
	}

	ingressPath, err := url.JoinPath(subPath, appId)
	if err != nil {
		return "", errors.WithMessage(err, "failed to join ingress path")
	}

	return ingressPath, nil
}
