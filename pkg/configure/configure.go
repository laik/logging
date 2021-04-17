package configure

import (
	"fmt"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/datasource/k8s"
	client "k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type RuntimeMode string

var AppRuntimeMode RuntimeMode = DEFAULT

func SetTheAppRuntimeMode(rm RuntimeMode) {
	AppRuntimeMode = rm
}

const (
	// InCluster when deploying in k8s, use this option
	INCLUSTER RuntimeMode = "InCluster"
	// Default when deploying in non k8s, use this option and the is default option
	DEFAULT RuntimeMode = "Default"
)

// InstallConfigure ...
type InstallConfigure struct {
	// kubernetes reset config
	RestConfig *rest.Config
	// k8s CacheInformerFactory
	*k8s.CacheInformerFactory
	// k8s client
	client.Interface
	// ResourceLister resource lister
	k8s.ResourceLister

	*kubernetes.Clientset
	//*rest.RESTClient
}

func NewInstallConfigure(k8sResLister k8s.ResourceLister) (*InstallConfigure, error) {
	var (
		cli         client.Interface
		resetConfig *rest.Config
		err         error
	)

	switch AppRuntimeMode {
	case DEFAULT:
		fmt.Printf("%s start app is Default mode", common.INFO)
		cli, resetConfig, err = k8s.BuildClientSet(*common.KubeConfig)
	case INCLUSTER:
		fmt.Printf("%s start app is InCluster mode", common.INFO)
		_, resetConfig, err = k8s.CreateInClusterConfig()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("not define the runtime mode")
	}

	if resetConfig == nil {
		return nil, fmt.Errorf("failed to init rest config")
	}

	cacheInformerFactory, err := k8s.NewCacheInformerFactory(k8sResLister, resetConfig)
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(resetConfig)

	return &InstallConfigure{
		CacheInformerFactory: cacheInformerFactory,
		Interface:            cli,
		RestConfig:           resetConfig,
		ResourceLister:       k8sResLister,
		Clientset:            clientSet,
	}, nil
}

func init() {
	if os.Getenv("INCLUSTER") != "" {
		SetTheAppRuntimeMode(INCLUSTER)
	}
}
