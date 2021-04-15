package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yametech/logging/pkg/api"
	"github.com/yametech/logging/pkg/configure"
	"github.com/yametech/logging/pkg/datasource"
	"github.com/yametech/logging/pkg/datasource/k8s"
	"github.com/yametech/logging/pkg/service"
	"github.com/yametech/logging/pkg/types"
)

var ns string

func main() {
	flag.StringVar(&ns, "ns", "", "-ns kube-system")
	flag.Parse()

	if ns == "" {
		ns = os.Getenv("NAMESPACE")
		if ns == "" {
			panic("ns must not be empty")
		}
	}

	config, err := configure.NewInstallConfigure(k8s.NewResources(
		[]string{},
		types.KubernetesResourceInit,
		types.YameCloudResourceInit,
	))
	if err != nil {
		panic(fmt.Sprintf("new install configure error %s", err))
	}
	server := api.NewServer("0.0.0.0:9999", ns, service.NewIService(datasource.NewIDataSource(config)))
	if err := server.Start(); err != nil {
		panic(err)
	}
}
