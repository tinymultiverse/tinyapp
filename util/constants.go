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

const AnyUserName = "anyuser"

const (
	TLSSecretVolumeName = "tls-secret"
	TLSSecretMountPath  = "/tls-secret"
	TLSKeyFile          = TLSSecretMountPath + "/tls.key"
	TLSCertFile         = TLSSecretMountPath + "/tls.crt"
)

const (
	K8sNameLabel       = "app.kubernetes.io/name"
	K8sPartOfLabel     = "app.kubernetes.io/part-of"
	TinyAppPartOfLabel = "tinyapp"
)

const (
	K8sMaxNameSize = 253
)

const (
	Https = "https://"
	Http  = "http://"
)
