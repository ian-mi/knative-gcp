/*
Copyright 2019 Google LLC.

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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Stackdriver is a specification for a Stackdriver Source resource
type Stackdriver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackdriverSpec   `json:"spec"`
	Status StackdriverStatus `json:"status"`
}

var (
	_ runtime.Object = (*Stackdriver)(nil)
)

type StackdriverSpec struct {
	// This brings in the PubSub based Source Specs. Includes:
	// Sink, CloudEventOverrides, Secret, PubSubSecret, and Project
	duckv1alpha1.PubSubSpec

	// Optional log filter to apply. If unsepcified, will
	// subscribe to all log events.  See
	// https://cloud.google.com/logging/docs/view/advanced-queries
	Filter string `json:"filter"`
}

type StackdriverStatus struct {
	duckv1alpha1.PubSubStatus

	SinkID string
}

// GetGroupVersionKind ...
func (stackdriver *Stackdriver) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Stackdriver")
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StackdriverList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Stackdriver `json"items"`
}
