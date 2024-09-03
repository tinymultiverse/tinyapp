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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tinymultiverse/tinyapp/controller/internal"

	"github.com/tinymultiverse/tinyapp/controller/util"
	globalutil "github.com/tinymultiverse/tinyapp/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	GitSyncBranchEnvVarName    = "GIT_SYNC_BRANCH"
	GitSyncTagEnvVarName       = "GIT_SYNC_REV"
	MainFileEnvVarName         = "MAIN_FILE"
	MainDirectoryEnvVarName    = "BASE_DIR"
	RequirementsFileEnvVarName = "REQUIREMENTS_FILE"
	TinyAppTypeEnvVarName      = "TINY_APP_TYPE"
	TinyAppNameEnvVarName      = "TINY_APP_NAME"
	StreamlitPortEnvVarName    = "STREAMLIT_PORT"
	DashPortEnvVarName         = "DASH_PORT"
	StreamlitBaseUrlEnvVarName = "STREAMLIT_BASE_URL"
	DashBaseUrlEnvVarName      = "DASH_URL_BASE_PATHNAME"
	RequirementsFileName       = "requirements.txt"
)

func BuildDeployment(app *v1alpha1.TinyApp, env internal.EnvVars) (*appsv1.Deployment, error) {
	var volumes []corev1.Volume
	for _, volumeClaim := range app.Spec.VolumeClaims {
		volumes = append(volumes, buildVolume(volumeClaim.Name))
	}

	// Add volume for TLS secret
	if env.GatewayMetricsTlsEnabled {
		volumes = append(volumes, corev1.Volume{
			Name: globalutil.TLSSecretVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: env.TLSSecretName,
				},
			},
		})
	}

	appContainer, err := buildAppContainer(app, env)
	if err != nil {
		return nil, err
	}

	containers := []corev1.Container{appContainer, buildGatewayContainer(app, env)}
	var initContainers []corev1.Container

	if app.Spec.SourceType == v1alpha1.SourceTypeGit {
		initContainers = append(initContainers, buildGitSyncContainer(app, env))
		volumes = append(volumes, buildGitCloneVolume())
		volumes = append(volumes, buildGitTokenVolume(app.Spec.GitConfig.TokenSecretName))
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            app.Name,
			Namespace:       env.TinyAppNamespace,
			Labels:          app.Labels,
			OwnerReferences: createOwnerRefs(app),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: app.Labels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      app.GetLabels(),
					Annotations: env.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: env.AppServiceAccount,
					InitContainers:     initContainers,
					Containers:         containers,
					Volumes:            volumes,
				},
			},
		},
	}

	hash, err := getObjectHash(deployment)
	if err != nil {
		return nil, err
	}
	deployment.Annotations = map[string]string{
		util.AnnotationResourceHash: hash,
	}

	return deployment, nil
}

func buildAppContainer(app *v1alpha1.TinyApp, env internal.EnvVars) (corev1.Container, error) {
	if strings.TrimSpace(app.Spec.Image) == "" {
		return corev1.Container{}, errors.New("image name is empty")
	}

	envVars, err := buildAppEnvVars(app, env)
	if err != nil {
		return corev1.Container{}, err
	}

	appContainer := corev1.Container{
		Name:            util.AppContainerName,
		Image:           app.Spec.Image,
		ImagePullPolicy: corev1.PullAlways,
		Args: []string{
			"start-tinyapp.py",
		},
		Env: envVars,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				// TODO make it part of app spec
				corev1.ResourceCPU:    resource.MustParse(util.AppCPURequest),
				corev1.ResourceMemory: resource.MustParse(util.AppMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(util.AppCPULimit),
				corev1.ResourceMemory: resource.MustParse(util.AppMemoryLimit),
			},
		},
		VolumeMounts: buildAppVolumeMounts(app),
	}

	return appContainer, nil
}

func buildAppEnvVars(app *v1alpha1.TinyApp, env internal.EnvVars) ([]corev1.EnvVar, error) {
	mainFileDir, mainFileName := filepath.Split(app.Spec.MainFilePath)

	baseDir := util.GitRootDir + "/" + util.GitDestDir
	if app.Spec.SourceType == v1alpha1.SourceTypeFileSystem {
		baseDir = getMainVolumeClaimMountPath(app)
		if baseDir == "" {
			return nil, errors.New("main volume claim name not found")
		}
	}

	baseUrl, err := BuildIngressPath(app.Spec.IngressSubPath, app.Name)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to build ingress path")
	}

	appEnvVars := []corev1.EnvVar{
		{
			Name:  MainFileEnvVarName,
			Value: mainFileName,
		},
		{
			Name:  MainDirectoryEnvVarName,
			Value: filepath.Join(baseDir, mainFileDir),
		},
		{
			Name:  TinyAppTypeEnvVarName,
			Value: strings.ToLower(string(app.Spec.AppType)),
		},
		{
			Name:  TinyAppNameEnvVarName,
			Value: app.Name,
		},
		{
			Name:  RequirementsFileEnvVarName,
			Value: RequirementsFileName,
		},
		{
			Name:  StreamlitPortEnvVarName,
			Value: util.DefaultAppPort,
		},
		{
			Name:  DashPortEnvVarName,
			Value: util.DefaultAppPort,
		},
		{
			Name:  StreamlitBaseUrlEnvVarName,
			Value: baseUrl,
		},
		{
			Name: DashBaseUrlEnvVarName,
			// Forward slash at the end is required by gunicorn
			Value: baseUrl + "/",
		},
	}

	for key, val := range env.DefaultAppEnvVars {
		appEnvVars = addEnvVar(key, val, appEnvVars)
	}

	for _, envVar := range app.Spec.EnvVars {
		appEnvVars = addEnvVar(envVar.Name, envVar.Value, appEnvVars)
	}

	return appEnvVars, nil
}

// addEnvVar adds key/value env var to envVars; if key already exists, it updates the value.
func addEnvVar(key, value string, envVars []corev1.EnvVar) []corev1.EnvVar {
	for index, envVar := range envVars {
		if envVar.Name == key {
			envVars[index].Value = value
			return envVars
		}
	}

	return append(envVars, corev1.EnvVar{Name: key, Value: value})
}

// buildAppVolumeMounts returns a list of volume mounts for app container.
func buildAppVolumeMounts(app *v1alpha1.TinyApp) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount
	for _, volumeClaim := range app.Spec.VolumeClaims {
		volumeMounts = append(volumeMounts, buildVolumeMount(*volumeClaim))
	}

	// Volume mount for git sync to clone the git repository
	if app.Spec.SourceType == v1alpha1.SourceTypeGit {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      util.GitCloneVolumeName,
			MountPath: util.GitRootDir,
		})
	}

	return volumeMounts
}

// getMainVolumeClaimMountPath returns the mount path of the main volume claim.
func getMainVolumeClaimMountPath(app *v1alpha1.TinyApp) string {
	for _, volumeClaim := range app.Spec.VolumeClaims {
		if volumeClaim.Name == app.Spec.MainVolumeClaimName {
			return volumeClaim.MountPath
		}
	}

	return ""
}

func buildGatewayContainer(app *v1alpha1.TinyApp, env internal.EnvVars) corev1.Container {
	envs := buildGatewayEnvVars(app, env)

	gatewayContainer := corev1.Container{
		Name:  util.GatewayContainerName,
		Image: env.GatewayImage,
		Ports: []corev1.ContainerPort{
			{
				Protocol:      "TCP",
				ContainerPort: util.DefaultGatewayPort,
				Name:          "gatewayport",
			},
		},
		ImagePullPolicy: corev1.PullAlways,
		Env:             envs,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(util.GatewayContainerCPURequest),
				corev1.ResourceMemory: resource.MustParse(util.GatewayContainerMemoryRequest),
			},
			Limits: corev1.ResourceList{
				// TODO read from env
				corev1.ResourceCPU:    resource.MustParse(util.GatewayContainerCPULimit),
				corev1.ResourceMemory: resource.MustParse(util.GatewayContainerMemoryLimit),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      globalutil.TLSSecretVolumeName,
				MountPath: globalutil.TLSSecretMountPath,
				ReadOnly:  true,
			},
		},
	}

	return gatewayContainer
}

func buildGatewayEnvVars(app *v1alpha1.TinyApp, env internal.EnvVars) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{Name: "HTTP_PORT", Value: strconv.Itoa(int(util.DefaultGatewayPort))},
		{Name: "TINY_APP_NAME", Value: app.Name},
		{Name: "METRICS_ENABLED", Value: strconv.FormatBool(env.GatewayMetricsEnabled)},
		{Name: "METRICS_TLS_ENABLED", Value: strconv.FormatBool(env.GatewayMetricsTlsEnabled)},
		{Name: "METRICS_PORT", Value: env.GatewayMetricsPort},
		{Name: "METRICS_PATH", Value: env.GatewayMetricsPath},
	}

	envVars = append(envVars, buildEnvVarsList(env.GatewayEnvVars)...)

	return envVars
}

func buildGitSyncContainer(app *v1alpha1.TinyApp, env internal.EnvVars) corev1.Container {
	gitConfig := app.Spec.GitConfig

	var gitRefEnvVarName string
	if gitConfig.IsTag {
		gitRefEnvVarName = GitSyncTagEnvVarName
	} else {
		gitRefEnvVarName = GitSyncBranchEnvVarName
	}

	gitSyncEnvVars := []corev1.EnvVar{
		{Name: "GIT_SYNC_ROOT", Value: util.GitRootDir},
		{Name: "GIT_SYNC_DEST", Value: util.GitDestDir},
		{Name: "GIT_SYNC_PASSWORD_FILE", Value: util.GitTokenFilePath},
		{Name: "GIT_SYNC_MAX_SYNC_FAILURES", Value: "-1"},
		{Name: "GIT_SYNC_USERNAME", Value: "abc"}, // Can be anything except empty space
		{Name: "GIT_SYNC_REPO", Value: gitConfig.GitUrl},
		{Name: "GIT_SYNC_ONE_TIME", Value: "true"},
		{Name: "GIT_SYNC_WAIT", Value: "0"},
		{Name: gitRefEnvVarName, Value: gitConfig.GitRef},
	}

	extraEnvVars := buildEnvVarsList(env.GitSyncEnvVars)

	container := corev1.Container{
		Name:            util.GitSyncContainerName,
		Image:           env.GitSyncImage,
		ImagePullPolicy: corev1.PullAlways,
		Env:             append(gitSyncEnvVars, extraEnvVars...),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(util.GitSyncContainerCPURequest),
				corev1.ResourceMemory: resource.MustParse(util.GitSyncContainerMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(util.GitSyncContainerCPULimit),
				corev1.ResourceMemory: resource.MustParse(util.GitSyncContainerMemoryLimit),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      util.GitCloneVolumeName,
				MountPath: util.GitRootDir,
			},
			{
				Name:      util.GitTokenVolumeName,
				MountPath: util.GitTokenFilePath,
				SubPath:   util.GitTokenSecretKey,
			},
		},
	}

	return container
}
