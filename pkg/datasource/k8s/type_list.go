package k8s

import "k8s.io/apimachinery/pkg/runtime/schema"

const (

	// Kubernetes
	Namespace = "namespaces"
	Pod       = "pods"
)

func rsInit(rs *Resources) {
	// kubernetes
	rs.register(Namespace, schema.GroupVersionResource{Group: "", Version: "v1", Resource: Namespace})
	rs.register(Pod, schema.GroupVersionResource{Group: "", Version: "v1", Resource: Pod})
}
