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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretSpec defines the desired state of Secret
// +kubebuilder:validation:Required
type SecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Secret. Edit secret_types.go to remove/update
	Secret    SecretResource  `json:"secret"`
	Generator SecretGenerator `json:"generator"`
}

// +kubebuilder:validation:Required
type SecretResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// +kubebuilder:validation:Required
type SecretGenerator struct {
	Type   string `json:"type"`
	Length int    `json:"length"`
}

// SecretStatus defines the observed state of Secret
type SecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Secret is the Schema for the secrets API
type Secret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretSpec   `json:"spec,omitempty"`
	Status SecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretList contains a list of Secret
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Secret{}, &SecretList{})
}
