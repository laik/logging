package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="service",type=string,JSONPath=`.spec.service_name`
// +kubebuilder:printcolumn:name="filter",type=string,JSONPath=`.spec.filter`
// +kubebuilder:printcolumn:name="pod",type=string,JSONPath=`.spec.pod`
type SlackTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SlackTaskSpec   `json:"spec,omitempty"`
	Status            SlackTaskStatus `json:"status,omitempty"`
}

// OutputSpec defines the desired state of OutputSpec
type SlackTaskSpec struct {
	Type        watch.EventType `json:"type"`
	Ns          string          `json:"ns"`
	ServiceName string          `json:"service_name"`
	FilterName  string          `json:"filter"`
	Node        string          `json:"node"`
	Pod         string          `json:"pod"`
	Ips         []string        `json:"ips"`
	Offset      uint64          `json:"offset"`
}

// OutputStatus defines the observed state of OutputStatus
type SlackTaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

func init() {
	register(&SlackTask{})
}
