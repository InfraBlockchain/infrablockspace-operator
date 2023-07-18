/*
Copyright 2023.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InfraBlockSpaceSpec defines the desired state of InfraBlockSpace
type InfraBlockSpaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// app
	Region string `json:"region,omitempty"`
	// Foo is an example field of InfraBlockSpace. Edit infrablockspace_types.go to remove/update
	Image string `json:"image,binding:required"`

	// port
	Port Port `json:"port,omitempty"`

	// replicas
	Replicas int32 `json:"replicas,omitempty"`

	// chainSpec
	ChainSpec string `json:"chainSpec,omitempty"`

	// keys
	Keys *[]Key `json:"keys,omitempty"`

	// Periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// Actions that the management system should take in response to container lifecycle events.
	// +optional
	Lifecycle *corev1.Lifecycle            `json:"lifecycle,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type Port struct {
	RPCPort int32 `json:"rpcPort,omitempty"`
	WSPort  int32 `json:"wsPort,omitempty"`
	P2PPort int32 `json:"p2pPort,omitempty"`
}

type Key struct {
	KeyType string `json:"type,omitempty"`
	Scheme  string `json:"scheme,omitempty"`
	Seed    string `json:"seed,omitempty"`
}

// InfraBlockSpaceStatus defines the observed state of InfraBlockSpace
type InfraBlockSpaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SecretStatus string
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InfraBlockSpace is the Schema for the infrablockspaces API
type InfraBlockSpace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InfraBlockSpaceSpec   `json:"spec,omitempty"`
	Status InfraBlockSpaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InfraBlockSpaceList contains a list of InfraBlockSpace
type InfraBlockSpaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfraBlockSpace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InfraBlockSpace{}, &InfraBlockSpaceList{})
}
