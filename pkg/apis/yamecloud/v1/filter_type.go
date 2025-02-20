package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="max_length",type=string,JSONPath=`.spec.max_length`
// +kubebuilder:printcolumn:name="expr",type=string,JSONPath=`.spec.expr`

type Filter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FilterSpec   `json:"spec,omitempty"`
	Status            FilterStatus `json:"status,omitempty"`
}

// OutputSpec defines the desired state of OutputSpec
type FilterSpec struct {
	MaxLength uint64 `json:"max_length,omitempty"`
	Expr      string `json:"expr,omitempty"`
}

// OutputStatus defines the observed state of OutputStatus
type FilterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

func init() {
	register(&Filter{})
}
