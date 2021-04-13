package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="selector",type=string,JSONPath=`.spec.selector`
// +kubebuilder:printcolumn:name="all_tasks",type=string,JSONPath=`.status.all_tasks`
// +kubebuilder:resource:shortName=slacks
type Slack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlackSpec   `json:"spec,omitempty"`
	Status SlackStatus `json:"status,omitempty"`
}

type Task struct {
	Ns          string `json:"ns"`
	ServiceName string `json:"service_name"`
	Filter      Filter `json:"filter"`
	Pods        []Pod  `json:"pods"`
}

type Pod struct {
	Node      string   `json:"node"`
	Pod       string   `json:"pod"`
	Container string   `json:"container"`
	Ips       []string `json:"ips"`
	Offset    int      `json:"offset"`
}

type Filter struct {
	MaxLength string `json:"max_length,omitempty"`
	Expr      string `json:"expr,omitempty"`
}

type SlackSpec struct {
	Selector string `json:"selector,omitempty"`
	// AddTasks that need to be collected
	AddTasks []Task `json:"add_tasks,omitempty"`
	// DeleteTasks
	DeleteTasks []Task   `json:"delete_tasks,omitempty"`
	AllTasks    []Record `json:"all_tasks,omitempty"`
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
	Filter      `json:"filter"`
}

func (r *Record) ToPod() Pod {
	return Pod{
		Node:      r.NodeName,
		Pod:       r.PodName,
		Container: r.Container,
		Ips:       r.Ips,
		Offset:    r.Offset,
	}
}

type SlackStatus struct {
}

func init() {
	register(&Slack{})
}
