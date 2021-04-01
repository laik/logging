/*
Copyright 2021 yametech Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OutputType string

const (
	KAFKA OutputType = "kafka"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Output struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Spec            OutputSpec   `json:"spec,omitempty"`
	Status          OutputStatus `json:"status,omitempty"`
}

// OutputSpec defines the desired state of OutputSpec
type OutputSpec struct {
	Type    *OutputType `json:"type,omitempty"`
	Address *string     `json:"address,omitempty"`
}

// OutputStatus defines the observed state of OutputStatus
type OutputStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type OutputList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Output `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LoggingTask struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Spec   LoggingTaskSpec   `json:"spec,omitempty"`
	Status LoggingTaskStatus `json:"status,omitempty"`
}

type Task struct {
	PodName string `json:"pod_name,omitempty"`
}

type LoggingTaskSpec struct {
	AddTasks    []Task `json:"add_tasks,omitempty"`
	DeleteTasks []Task `json:"delete_tasks,omitempty"`
	AllTasks    []Task `json:"all_tasks,omitempty"`
}

type LoggingTaskStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LoggingTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoggingTask `json:"items"`
}

func init() {
	register(&Output{}, &OutputList{})
	register(&LoggingTask{}, &LoggingTaskList{})
}
