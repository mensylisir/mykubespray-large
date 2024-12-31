package kubernetes

import (
	"fmt"
	helm "github.com/mittwald/go-helm-client"
	"github.com/pkg/errors"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/httpx"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"time"
)

type K8sClient struct {
	Clientset       *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
	CRDClient       *apiextensionsclientset.Clientset
	HelmClient      helm.Client
	HttpClient      *http.Client
	InformerFactory informers.SharedInformerFactory
}

func NewK8sClient(config entity.K8sConfig) (*K8sClient, error) {
	kubeconf, err := GetKubernetesRestConfig(&config)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(kubeconf)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(kubeconf)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeconf)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %v", err)
	}
	crdclient, err := apiextensionsclientset.NewForConfig(kubeconf)
	if err != nil {
		return nil, fmt.Errorf("failed to create crd client: %v", err)
	}
	httpClient := &http.Client{Timeout: 10 * time.Second, Transport: &httpx.CustomTransport{}}
	//helmClient, err := NewHelmClient(&config)
	helmClient, err := NewHelmClientFromRestConfig(kubeconf)
	if err != nil {
		return nil, fmt.Errorf("failed to create helm client: %v", err)
	}
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)

	return &K8sClient{
		Clientset:       clientset,
		DynamicClient:   dynamicClient,
		DiscoveryClient: discoveryClient,
		CRDClient:       crdclient,
		HelmClient:      helmClient,
		HttpClient:      httpClient,
		InformerFactory: informerFactory,
	}, nil
}

func GetKubernetesRestConfig(c *entity.K8sConfig) (*rest.Config, error) {
	var kubeConf *rest.Config
	var err error
	if c.KubeconfigPath != "" {
		data, err := os.ReadFile(c.KubeconfigPath)
		if err != nil {
			return nil, err
		}
		c.Kubeconfig = string(data)
		kubeConf, err = clientcmd.BuildConfigFromFlags("", c.KubeconfigPath)
		if err != nil {
			return nil, err
		}
	} else if c.Kubeconfig != "" {
		kubeConf, err = clientcmd.RESTConfigFromKubeConfig([]byte(c.Kubeconfig))
		if err != nil {
			return nil, err
		}
	} else if c.ApiServer != "" && c.Token != "" {
		kubeConf = &rest.Config{
			Host:        c.ApiServer,
			BearerToken: c.Token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	} else {
		kubeConf, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	return kubeConf, nil
}

func NewHelmClient(c *entity.K8sConfig) (helm.Client, error) {
	options := &helm.KubeConfClientOptions{
		Options: &helm.Options{
			Namespace:        "",
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true,
			DebugLog: func(format string, v ...interface{}) {
				logger.GetLogger().Printf(format, v...)
			},
		},
		KubeContext: "",
		KubeConfig:  []byte(c.Kubeconfig),
	}
	client, err := helm.NewClientFromKubeConf(options, helm.Burst(100), helm.Timeout(10e9))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewHelmClientFromRestConfig(kubeConf *rest.Config) (helm.Client, error) {
	options := &helm.RestConfClientOptions{
		Options: &helm.Options{
			Namespace:        "", // Change this to the namespace you wish to install the chart in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			DebugLog: func(format string, v ...interface{}) {
				logger.GetLogger().Printf(format, v...)
			},
		},
		RestConfig: kubeConf,
	}
	client, err := helm.NewClientFromRestConf(options)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetKubernetesRestConfigFromKubeConfig(c *entity.K8sConfig) (*rest.Config, error) {
	kubeConf, err := clientcmd.RESTConfigFromKubeConfig([]byte(c.Kubeconfig))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("new kubernetes client with config failed: %v", err))
	}
	return kubeConf, nil
}

func NewKubernetesClient(c *entity.K8sConfig) (*kubernetes.Clientset, error) {
	kubeConf, err := GetKubernetesRestConfig(c)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return client, errors.Wrap(err, fmt.Sprintf("new kubernetes client with config failed: %v", err))
	}
	return client, nil
}

func NewKubernetesDynamicClient(c *entity.K8sConfig) (dynamic.Interface, error) {
	kubeConf, err := GetKubernetesRestConfig(c)
	if err != nil {
		return nil, err
	}
	dynamicClient, err1 := dynamic.NewForConfig(kubeConf)
	if err1 != nil {
		return nil, errors.Wrap(err1, fmt.Sprintf("new kubernetes dynamic client with config failed: %v", err1))
	}
	return dynamicClient, nil
}

func NewKubernetesDiscoveryClient(c *entity.K8sConfig) (*discovery.DiscoveryClient, error) {
	client, err := NewKubernetesClient(c)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("new kubernetes dynamic client with config failed: %v", err))
	}
	discoveryClient := discovery.NewDiscoveryClient(client.RESTClient())
	return discoveryClient, nil
}
