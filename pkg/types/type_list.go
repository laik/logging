package types

import (
	"github.com/yametech/logging/pkg/datasource/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	YameCloudApiGroup = "yamecloud.io"

	// Kubernetes
	Namespace = "namespaces"
	Pod       = "pods"

	// Yamecloud Logging
	Output      = "outputs"
	LoggingTask = "loggingtasks"
)

func KubernetesResourceInit(rs *k8s.Resources) {
	// kubernetes
	rs.Registry(Namespace, schema.GroupVersionResource{Group: "", Version: "v1", Resource: Namespace})
	rs.Registry(Pod, schema.GroupVersionResource{Group: "", Version: "v1", Resource: Pod})

}

func YameCloudResourceInit(rs *k8s.Resources) {
	// yamecloud logging operator resource
	rs.Registry(Output, schema.GroupVersionResource{Group: YameCloudApiGroup, Version: "v1", Resource: Output})
	rs.Registry(LoggingTask, schema.GroupVersionResource{Group: YameCloudApiGroup, Version: "v1", Resource: LoggingTask})
}
