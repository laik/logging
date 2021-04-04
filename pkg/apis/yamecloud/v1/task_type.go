package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LoggingTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoggingTaskSpec   `json:"spec,omitempty"`
	Status LoggingTaskStatus `json:"status,omitempty"`
}

type Task struct {
	ServiceName string   `json:"service_name,omitempty"`
	Ns          string   `json:"ns,omitempty"`
	PodName     string   `json:"pod_name,omitempty"`
	Ips         []string `json:"ips,omitempty"`
	Output      Output   `json:"output,omitempty"`
	Node        string   `json:"node,omitempty"`
	Rules       []Rule   `json:"rules,omitempty"`
}

type Rule struct {
	MaxLength  uint64 `json:"max_length,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type LoggingTaskSpec struct {
	AddTasks    map[string]Task `json:"add_tasks,omitempty"`
	DeleteTasks map[string]Task `json:"delete_tasks,omitempty"`
	AllTasks    map[string]Task `json:"all_tasks,omitempty"`
}

type LoggingTaskStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LoggingTaskList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []LoggingTask `json:"items"`
}

func init() {
	register(&LoggingTask{}, &LoggingTaskList{})
}
