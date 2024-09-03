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

package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:subresource:status
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TinyApp is a K8s resource corresponding to an application for which a user
// can spin up TinyAppInstances of
type TinyApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinyAppSpec   `json:"spec"`
	Status TinyAppStatus `json:"status,omitempty"`
}

type TinyAppSpec struct {
	// DisplayName is the non-unique name of app.
	DisplayName string `json:"displayName"`
	// Description is short description of app.
	Description string `json:"description" required:"false"`
	// Documentation holds url to TinyApp documentation.
	Documentation string `json:"documentation"`
	// Image is Docker image used to run the app.
	Image string `json:"image"`
	// AppType is type of app.
	AppType `json:"appType"`
	// SourceType determines where app code comes from (ex. git).
	SourceType `json:"sourceType"`
	// GitConfig contains git repository information.
	// Used & required only when SourceType is Git.
	*GitConfig `json:"gitConfig"`
	// MainFilePath is the path to app main file, relative to base directory.
	// Base directory will be /app when SourceType is Git and main volumeClaim's mount path when SourceType is FileSystem.
	// Parent directory of main file will be the working directory (cwd) for the app process.
	MainFilePath string `json:"mainFile"`
	// EnvVars is environment variables to set in app.
	EnvVars []*corev1.EnvVar `json:"envVars"`
	// Volume claims to mount in app.
	VolumeClaims []*VolumeClaim `json:"volumeClaims"`
	// MainVolumeClaimName is the name of the volume claim that contains the main file.
	// Used & required only when SourceType is FileSystem.
	MainVolumeClaimName string `json:"mainVolumeClaimName"`
	// IngressDomain is domain of the ingress.
	IngressDomain string `json:"ingressDomain"`
	// IngressSubPath is the path to the app in the ingress.
	IngressSubPath string `json:"ingressPath"`
	// IngressTlsEnabled specified whether Tls is enabled for ingress.
	IngressTlsEnabled bool `json:"ingressTlsEnabled"`
}

type AppType string

// App types
const (
	AppTypeStreamlit AppType = "Streamlit"
	AppTypeDash      AppType = "Dash"
	AppTypeUnknown   AppType = "Unknown"
)

type SourceType string

const (
	// SourceTypeGit means the app source code is coming from a git repository
	SourceTypeGit SourceType = "Git"
	// SourceTypeFileSystem means the app source code is coming from mounted file system
	SourceTypeFileSystem SourceType = "FileSystem"
	// SourceTypeUnknown means the app source type is unknown
	SourceTypeUnknown = "Unknown"
)

type GitConfig struct {
	// Url to clone git repo when app source type is git
	GitUrl string `json:"gitUrl"`
	// If true, GitRef refers to tag instead of branch name
	IsTag bool `json:"isTag"`
	// Branch or tag name depending on isTag
	GitRef string `json:"gitRef"`
	// Secret name containing git token. Must be in the same namespace as TinyApp object.
	TokenSecretName string `json:"tokenSecretName"`
}

type Volume struct {
	ClaimName string `json:"claimName"`
}

type VolumeClaim struct {
	Name      string `json:"name"`
	SubPath   string `json:"subPath"`
	MountPath string `json:"mountPath"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TinyAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TinyApp `json:"items"`
}

type TinyAppPhase string

const (
	// TinyAppDeployed means the underlying resources have been created successfully.
	TinyAppDeployed TinyAppPhase = "Deployed"
	// TinyAppFailed means that some or all of the underlying resources failed to be set up successfully.
	TinyAppFailed TinyAppPhase = "Failed"
)

type TinyAppConditionType string

const (
	DeploymentCreated TinyAppConditionType = "DeploymentCreated"
	ServiceCreated    TinyAppConditionType = "ServiceCreated"
	IngressCreated    TinyAppConditionType = "IngressCreated"
)

// TinyAppStatus defines the observed state of TinyApp
type TinyAppStatus struct {
	// +optional
	Phase TinyAppPhase `json:"phase"`
	// Important: Run "make" to regenerate code after modifying this file
	// Represents the latest available observations of a deployment's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []*Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type Condition struct {
	// Type of tracking server condition.
	Type TinyAppConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

func (s *TinyAppStatus) InitConditions() {
	if len(s.Conditions) > 0 {
		return // Already initialized
	}

	conditionTypes := []TinyAppConditionType{
		DeploymentCreated, ServiceCreated, IngressCreated,
	}

	var conditions []*Condition
	for _, ct := range conditionTypes {
		conditions = append(conditions, &Condition{
			Type:   ct,
			Status: v1.ConditionUnknown,
		})
	}

	s.Conditions = conditions
}

func (s *TinyAppStatus) SetConditionTrue(conditionType TinyAppConditionType) {
	for _, c := range s.Conditions {
		if c.Type == conditionType {
			if c.Status == v1.ConditionTrue {
				return
			}
			c.LastTransitionTime = metav1.NewTime(time.Now())
			c.Status = v1.ConditionTrue
		}
	}
}

func (s *TinyAppStatus) SetConditionFalseWithMessage(conditionType TinyAppConditionType, message string) {
	for _, c := range s.Conditions {
		if c.Type == conditionType {
			c.LastTransitionTime = metav1.NewTime(time.Now())
			c.Status = v1.ConditionFalse
			c.Message = message
		}
	}
}

func (s *TinyAppStatus) GetCondition(conditionType TinyAppConditionType) *Condition {
	for _, c := range s.Conditions {
		if c.Type == conditionType {
			return c
		}
	}
	return nil
}

func (s *TinyAppStatus) IsConditionTrue(conditionType TinyAppConditionType) bool {
	for _, c := range s.Conditions {
		if c.Type == conditionType {
			return c.Status == v1.ConditionTrue
		}
	}
	return false
}

func (s *TinyAppStatus) GetPhase() TinyAppPhase {
	for _, c := range s.Conditions {
		if c.Status != v1.ConditionTrue {
			return TinyAppFailed
		}
	}

	return TinyAppDeployed
}
