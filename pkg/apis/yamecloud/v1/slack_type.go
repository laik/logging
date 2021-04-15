package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="selector",type=string,JSONPath=`.spec.selector`
// +kubebuilder:printcolumn:name="records",type=string,JSONPath=`.status.records`
// +kubebuilder:resource:shortName=slacks
type Slack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlackSpec   `json:"spec,omitempty"`
	Status SlackStatus `json:"status,omitempty"`
}

type SlackSpec struct {
	Selector string   `json:"selector,omitempty"`
	Records  []Record `json:"records,omitempty"`
}

type Record struct {
	Container   string   `json:"container"`
	Ips         []string `json:"ips"`
	IsUpload    bool     `json:"is_upload"`
	LastOffset  int      `json:"last_offset"`
	NodeName    string   `json:"node_name"`
	Ns          string   `json:"ns"`
	Offset      int      `json:"offset"`
	Output      string   `json:"output"`
	Path        string   `json:"path"`
	PodName     string   `json:"pod_name"`
	ServiceName string   `json:"service_name"`
	State       string   `json:"state"`
	Filter      Filter   `json:"filter"`
}

type SlackStatus struct {
}

func init() {
	register(&Slack{})
}
