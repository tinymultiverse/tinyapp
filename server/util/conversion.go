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
	"github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	pb "github.com/tinymultiverse/tinyapp/pkg/server/api/v1/proto"
	"github.com/tinymultiverse/tinyapp/server/internal"
	globalutil "github.com/tinymultiverse/tinyapp/util"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConvertToProtoTinyApp converts k8s TinyApp to proto TinyApp.
func ConvertToProtoTinyApp(in *v1alpha1.TinyApp) (*pb.TinyApp, error) {
	if in == nil {
		return nil, errors.New("cannot convert nil to proto tiny app")
	}

	appUrl, err := GetURLForTinyApp(in.Spec.IngressDomain, in.Spec.IngressSubPath, in.Name, in.Spec.IngressTlsEnabled)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get tiny app url")
	}

	return &pb.TinyApp{
		AppRelease: &pb.TinyAppRelease{
			Id:                in.Name,
			AppUrl:            appUrl,
			CreationTimeStamp: in.CreationTimestamp.Time.String(),
		},
		AppDetail: &pb.TinyAppDetail{
			Name:                in.Spec.DisplayName,
			Description:         in.Spec.Description,
			Documentation:       in.Spec.Documentation,
			Image:               in.Spec.Image,
			AppType:             ConvertToProtoAppType(in.Spec.AppType),
			SourceType:          ConvertToProtoSourceType(in.Spec.SourceType),
			GitConfig:           ConvertToProtoGitConfig(in.Spec.GitConfig),
			MainFilePath:        in.Spec.MainFilePath,
			Env:                 ConvertToProtoEnvVars(in.Spec.EnvVars),
			VolumeClaims:        ConvertToProtoVolumeClaims(in.Spec.VolumeClaims),
			MainVolumeClaimName: in.Spec.MainVolumeClaimName,
		},
	}, nil
}

func ConvertToProtoEnvVars(envVars []*corev1.EnvVar) []*pb.EnvVar {
	var protoEnvVars []*pb.EnvVar
	for _, envVar := range envVars {
		protoEnvVar := &pb.EnvVar{
			Name:  envVar.Name,
			Value: envVar.Value,
		}
		protoEnvVars = append(protoEnvVars, protoEnvVar)
	}
	return protoEnvVars
}

func ConvertToProtoAppType(appType v1alpha1.AppType) pb.AppType {
	switch appType {
	case v1alpha1.AppTypeStreamlit:
		return pb.AppType_APP_TYPE_STREAM_LIT
	case v1alpha1.AppTypeDash:
		return pb.AppType_APP_TYPE_DASH
	default:
		return pb.AppType_APP_TYPE_UNKNOWN
	}
}

func ConvertToProtoSourceType(sourceType v1alpha1.SourceType) pb.SourceType {
	switch sourceType {
	case v1alpha1.SourceTypeGit:
		return pb.SourceType_SOURCE_TYPE_GIT
	case v1alpha1.SourceTypeFileSystem:
		return pb.SourceType_SOURCE_TYPE_FILE_SYSTEM
	default:
		return pb.SourceType_SOURCE_TYPE_UNKNOWN
	}
}

func ConvertToProtoGitConfig(gitConfig *v1alpha1.GitConfig) *pb.GitConfig {
	if gitConfig == nil {
		return nil
	}

	return &pb.GitConfig{
		Url:   gitConfig.GitUrl,
		Ref:   gitConfig.GitRef,
		IsTag: gitConfig.IsTag,
	}
}

func ConvertToProtoVolumeClaims(volumeClaims []*v1alpha1.VolumeClaim) []*pb.VolumeClaim {
	var protoVolumeClaims []*pb.VolumeClaim
	for _, volumeClaim := range volumeClaims {
		protoVolumeClaim := &pb.VolumeClaim{
			Name:      volumeClaim.Name,
			SubPath:   volumeClaim.SubPath,
			MountPath: volumeClaim.MountPath,
		}
		protoVolumeClaims = append(protoVolumeClaims, protoVolumeClaim)
	}
	return protoVolumeClaims
}

// ConvertToK8sTinyApp converts proto TinyAppDetail to k8s TinyApp.
func ConvertToK8sTinyApp(in *pb.TinyAppDetail, objName string, envVars internal.EnvVars) (*v1alpha1.TinyApp, error) {
	// Make sure input is not nil
	if in == nil {
		return nil, errors.New("empty TinyApp request")
	}

	image := envVars.DefaultAppImage
	if in.Image != "" {
		image = in.Image
	}

	tinyAppLabels := make(map[string]string)
	tinyAppLabels[globalutil.K8sNameLabel] = objName
	tinyAppLabels[globalutil.K8sPartOfLabel] = globalutil.TinyAppPartOfLabel

	return &v1alpha1.TinyApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:   objName,
			Labels: tinyAppLabels,
		},
		Spec: v1alpha1.TinyAppSpec{
			DisplayName:         in.Name,
			Description:         in.Description,
			Documentation:       in.Documentation,
			Image:               image,
			AppType:             ConvertToK8sAppType(in.AppType),
			SourceType:          ConvertToK8sSourceType(in.SourceType),
			GitConfig:           ConvertToK8sGitConfig(in.GitConfig, objName, envVars),
			MainFilePath:        in.MainFilePath,
			EnvVars:             ConvertToK8sEnvVars(in.Env),
			VolumeClaims:        ConvertToK8sVolumeClaims(in.VolumeClaims),
			MainVolumeClaimName: in.MainVolumeClaimName,
			IngressDomain:       envVars.AppIngressDomain,
			IngressSubPath:      envVars.AppIngressSubPath,
			IngressTlsEnabled:   envVars.AppIngressTlsEnabled,
		},
	}, nil
}

func ConvertToK8sEnvVars(envVars []*pb.EnvVar) []*corev1.EnvVar {
	var k8sEnvVars []*corev1.EnvVar
	for _, envVar := range envVars {
		k8sEnvVar := &corev1.EnvVar{
			Name:  envVar.Name,
			Value: envVar.Value,
		}
		k8sEnvVars = append(k8sEnvVars, k8sEnvVar)
	}
	return k8sEnvVars
}

func ConvertToK8sAppType(appType pb.AppType) v1alpha1.AppType {
	switch appType {
	case pb.AppType_APP_TYPE_STREAM_LIT:
		return v1alpha1.AppTypeStreamlit
	case pb.AppType_APP_TYPE_DASH:
		return v1alpha1.AppTypeDash
	default:
		return v1alpha1.AppTypeUnknown
	}
}

func ConvertToK8sSourceType(sourceType pb.SourceType) v1alpha1.SourceType {
	switch sourceType {
	case pb.SourceType_SOURCE_TYPE_GIT:
		return v1alpha1.SourceTypeGit
	case pb.SourceType_SOURCE_TYPE_FILE_SYSTEM:
		return v1alpha1.SourceTypeFileSystem
	default:
		return v1alpha1.SourceTypeUnknown
	}
}

func ConvertToK8sGitConfig(gitConfig *pb.GitConfig, appObjName string, envVars internal.EnvVars) *v1alpha1.GitConfig {
	if gitConfig == nil {
		return nil
	}

	tokenSecretName := envVars.DefaultGitTokenSecret
	if gitConfig.Token != "" {
		// If token is provided, we create a secret with name that is the same as tiny app object name.
		tokenSecretName = appObjName
	}

	return &v1alpha1.GitConfig{
		GitUrl:          gitConfig.Url,
		GitRef:          gitConfig.Ref,
		IsTag:           gitConfig.IsTag,
		TokenSecretName: tokenSecretName,
	}
}

func ConvertToK8sVolumeClaims(volumeClaims []*pb.VolumeClaim) []*v1alpha1.VolumeClaim {
	var k8sVolumeClaims []*v1alpha1.VolumeClaim
	for _, volumeClaim := range volumeClaims {
		k8sVolumeClaim := &v1alpha1.VolumeClaim{
			Name:      volumeClaim.Name,
			SubPath:   volumeClaim.SubPath,
			MountPath: volumeClaim.MountPath,
		}
		k8sVolumeClaims = append(k8sVolumeClaims, k8sVolumeClaim)
	}
	return k8sVolumeClaims
}
