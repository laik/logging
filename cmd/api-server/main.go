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
		panic("ns must not be empty")
	}

	config, err := configure.NewInstallConfigure(k8s.NewResources(
		[]string{},
		//types.KubernetesResourceInit,
		types.YameCloudResourceInit,
	))
	if err != nil {
		panic(fmt.Sprintf("new install configure error %s", err))
	}

	for err := range api.NewServer("0.0.0.0:9999", ns, service.NewIService(datasource.NewIDataSource(config))).Start() {
		fmt.Println(err)
		os.Exit(1)
	}

}
