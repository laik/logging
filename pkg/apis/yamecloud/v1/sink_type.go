package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type OutputType string

const (
	KAFKA OutputType = "kafka"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="address",type=string,JSONPath=`.spec.address`
// +kubebuilder:resource:shortName=outputs
type Sink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SinkSpec   `json:"spec,omitempty"`
	Status            SinkStatus `json:"status,omitempty"`
}

// OutputSpec defines the desired state of OutputSpec
type SinkSpec struct {
	Type      *OutputType `json:"type,omitempty"`
	Address   *string     `json:"address,omitempty"`
	Partition *int        `json:"partition,omitempty"`
}

// OutputStatus defines the observed state of OutputStatus
type SinkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

func init() {
	register(&Sink{})
}
