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

const (
	ControllerName = "tinyapp-controller"
)

const (
	AnnotationResourceHash = "resource-hash"
)

// Resource limits
const (
	AppCPURequest                 = "50m"
	AppCPULimit                   = "1"
	AppMemoryRequest              = "256Mi"
	AppMemoryLimit                = "3Gi"
	GatewayContainerCPURequest    = "50m"
	GatewayContainerCPULimit      = "250m"
	GatewayContainerMemoryRequest = "64Mi"
	GatewayContainerMemoryLimit   = "64Mi"
	GitSyncContainerCPURequest    = "50m"
	GitSyncContainerCPULimit      = "250m"
	GitSyncContainerMemoryRequest = "64Mi"
	GitSyncContainerMemoryLimit   = "256Mi"
)

const (
	AppContainerName     = "app"
	GitSyncContainerName = "git-sync"
	GatewayContainerName = "reverse-proxy"
)

const (
	DefaultGatewayPort int32 = 8080
	DefaultAppPort           = "5000"
)

const (
	GitCloneVolumeName = "git"
	GitRootDir         = "/app"
	GitDestDir         = "git-repo" // Relative to GitRootDir
	GitTokenVolumeName = "git-token"
	GitTokenSecretKey  = "token"
	GitTokenFilePath   = "/tmp/git-token.txt"
)
