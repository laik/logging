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

type Task struct {
	Ns          string `json:"ns"`
	ServiceName string `json:"service_name"`
	Filter      Filter `json:"filter"`
	Output      string `json:"output"`
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
	MaxLength uint64 `json:"max_length,omitempty"`
	Expr      string `json:"expr,omitempty"`
}

type SlackSpec struct {
	IsCollectAll  bool              `json:"collect_all,omitempty"`
	LabelSelector map[string]string `json:"label_selector"`

	AddTasks    map[string]Task `json:"add_tasks,omitempty"`
	DeleteTasks map[string]Task `json:"delete_tasks,omitempty"`
}

type PodResult struct {
	Node       string   `json:"node"`
	Pod        string   `json:"pod"`
	Container  string   `json:"container"`
	Ips        []string `json:"ips"`
	Offset     int      `json:"offset"`
	IsUploaded bool     `json:"is_uploaded,omitempty"`
	State      string   `json:"state"`
	Path       string   `json:"path"`
}

func (pr *PodResult) ToPod() Pod {
	return Pod{
		Node:      pr.Node,
		Pod:       pr.Pod,
		Container: pr.Container,
		Ips:       pr.Ips,
		Offset:    pr.Offset,
	}
}

type SlackStatus struct {
	AllTasks []PodResult `json:"all_tasks,omitempty"`
}

func init() {
	register(&Slack{})
}
