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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []LoggingTask `json:"items"`
}

func init() {
	register(&LoggingTask{}, &LoggingTaskList{})
}
