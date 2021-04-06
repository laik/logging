package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="add_tasks",type=string,JSONPath=`.spec.add_tasks`
// +kubebuilder:printcolumn:name="delete_tasks",type=string,JSONPath=`.spec.delete_tasks`
// +kubebuilder:printcolumn:name="all_tasks",type=string,JSONPath=`.spec.all_tasks`
// +kubebuilder:resource:shortName=slacks
type Slack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlackSpec   `json:"spec,omitempty"`
	Status SlackStatus `json:"status,omitempty"`
}

type Content struct {
	ServiceName string   `json:"service_name,omitempty"`
	Ns          string   `json:"ns,omitempty"`
	PodName     string   `json:"pod_name,omitempty"`
	Ips         []string `json:"ips,omitempty"`
	Output      string   `json:"output,omitempty"`
	Node        string   `json:"node,omitempty"`
	Rules       []Rule   `json:"rules,omitempty"`
}

type Rule struct {
	MaxLength  uint64 `json:"max_length,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type SlackSpec struct {
	AddTasks    map[string]Content `json:"add_tasks,omitempty"`
	DeleteTasks map[string]Content `json:"delete_tasks,omitempty"`
	AllTasks    map[string]Content `json:"all_tasks,omitempty"`
}

type SlackStatus struct {
}

func init() {
	register(&Slack{})
}
