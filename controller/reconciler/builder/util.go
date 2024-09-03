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
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/tinymultiverse/tinyapp/controller/util"
	"github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func createOwnerRefs(app *v1alpha1.TinyApp) []metav1.OwnerReference {
	return []metav1.OwnerReference{{
		APIVersion:         app.APIVersion,
		Kind:               app.Kind,
		Name:               app.GetName(),
		UID:                app.UID,
		Controller:         pointer.Bool(true),
		BlockOwnerDeletion: pointer.Bool(true),
	}}
}

func buildVolume(claimName string) corev1.Volume {
	return corev1.Volume{
		Name: claimName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	}
}

func buildVolumeMount(volumeClaim v1alpha1.VolumeClaim) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      volumeClaim.Name,
		SubPath:   volumeClaim.SubPath,
		MountPath: volumeClaim.MountPath,
	}
}

func buildGitCloneVolume() corev1.Volume {
	return corev1.Volume{
		Name: util.GitCloneVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func buildGitTokenVolume(secretName string) corev1.Volume {
	return corev1.Volume{
		Name: util.GitTokenVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}

}

func buildEnvVarsList(envVarsMap map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	if len(envVarsMap) == 0 {
		return envVars
	}

	for key, val := range envVarsMap {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: val})
	}

	return envVars
}

// hasher hashes a string
func hasher(value string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("%v", h.Sum32())
}

// getObjectHash returns hash of a given object
func getObjectHash(obj metav1.Object) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource")
	}
	return hasher(string(b)), nil
}
