/*
Copyright 2017 The Kubernetes Authors.

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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PassboltSecret is a specification for a PassboltSecret resource
type PassboltSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PassboltSecretSpec   `json:"spec"`
	Status PassboltSecretStatus `json:"status"`
}

// PassboltSecretSpec is the spec for a PassboltSecret resource
type PassboltSecretSpec struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
}

// PassboltSecretStatus is the status for a PassboltSecret resource
type PassboltSecretStatus struct {
	Created bool `json:"created"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PassboltSecretList is a list of PassboltSecret resources
type PassboltSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PassboltSecret `json:"items"`
}
