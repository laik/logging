package types

import (
	"github.com/yametech/logging/pkg/datasource/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	loggingYameCloudApiVersion = "logging.yamecloud.io"

	// Kubernetes
	Namespace = "namespaces"
	Pod       = "pods"

	// Yamecloud Logging
	Sink  = "sinks"
	Slack = "slacks"
)

func KubernetesResourceInit(rs *k8s.Resources) {
	// kubernetes
	rs.Registry(Namespace, schema.GroupVersionResource{Group: "", Version: "v1", Resource: Namespace})
	rs.Registry(Pod, schema.GroupVersionResource{Group: "", Version: "v1", Resource: Pod})

}

func YameCloudResourceInit(rs *k8s.Resources) {
	// yamecloud logging operator resource
	rs.Registry(Sink, schema.GroupVersionResource{Group: loggingYameCloudApiVersion, Version: "v1", Resource: Sink})
	rs.Registry(Slack, schema.GroupVersionResource{Group: loggingYameCloudApiVersion, Version: "v1", Resource: Slack})
}
