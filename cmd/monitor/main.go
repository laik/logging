package main

import (
	"flag"
	"fmt"
	"github.com/yametech/logging/pkg/configure"
	"github.com/yametech/logging/pkg/controller"
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
		panic("ns must not be empty")
	}

	config, err := configure.NewInstallConfigure(k8s.NewResources([]string{},
		types.YameCloudResourceInit,
		types.KubernetesResourceInit,
	))
	if err != nil {
		panic(fmt.Sprintf("new install configure error %s", err))
	}

	for err := range controller.NewPodMonitorController(
		ns,
		service.NewIService(datasource.NewIDataSource(config)),
	).Start() {
		panic(err)
	}

}
