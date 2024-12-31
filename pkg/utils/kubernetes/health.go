package kubernetes

import (
	"k8s.io/client-go/informers"
	"log"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type K8sHealthChecker struct {
	clientset       *kubernetes.Clientset
	informerFactory informers.SharedInformerFactory
	k8sclient       K8sClient
	cacheMutex      sync.RWMutex
	cacheData       *ClusterHealthData
	stopCh          chan struct{}
}

type ClusterHealthData struct {
	NodeHealth  map[string]bool
	PodHealth   map[string]bool
	LastChecked time.Time
}

func NewK8sHealthChecker(clientset *kubernetes.Clientset) *K8sHealthChecker {
	informerFactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)

	return &K8sHealthChecker{
		clientset:       clientset,
		informerFactory: informerFactory,
		cacheData:       &ClusterHealthData{NodeHealth: make(map[string]bool), PodHealth: make(map[string]bool)},
		stopCh:          make(chan struct{}),
	}
}

func (h *K8sHealthChecker) StartInformers() {
	h.informerFactory.Start(h.stopCh)
	h.informerFactory.WaitForCacheSync(h.stopCh)
}

func (h *K8sHealthChecker) CheckNodeHealth() {
	nodeLister := h.informerFactory.Core().V1().Nodes().Lister()

	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		log.Printf("failed to list nodes: %v", err)
		return
	}

	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	for _, node := range nodes {
		healthy := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				healthy = true
				break
			}
		}
		h.cacheData.NodeHealth[node.Name] = healthy
	}
}

func (h *K8sHealthChecker) CheckPodHealth(namespaces []string) {
	var wg sync.WaitGroup

	for _, namespace := range namespaces {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()
			podLister := h.informerFactory.Core().V1().Pods().Lister()

			pods, err := podLister.Pods(ns).List(labels.Everything())
			if err != nil {
				log.Printf("failed to list pods in namespace %s: %v", ns, err)
				return
			}

			h.cacheMutex.Lock()
			defer h.cacheMutex.Unlock()

			for _, pod := range pods {
				healthy := pod.Status.Phase == "Running"
				h.cacheData.PodHealth[pod.Name] = healthy
			}
		}(namespace)
	}

	wg.Wait()
}

func (h *K8sHealthChecker) UpdateCache() {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	h.cacheData.LastChecked = time.Now()

	go h.CheckNodeHealth()
	go h.CheckPodHealth([]string{"default", "kube-system"}) // Example namespaces
}

func (h *K8sHealthChecker) GetCachedHealthData() *ClusterHealthData {
	h.cacheMutex.RLock()
	defer h.cacheMutex.RUnlock()

	return h.cacheData
}

func (h *K8sHealthChecker) IsClusterHealthy() bool {
	h.UpdateCache()

	cacheData := h.GetCachedHealthData()

	healthy := true
	for _, isHealthy := range cacheData.NodeHealth {
		if !isHealthy {
			healthy = false
			break
		}
	}

	for _, isHealthy := range cacheData.PodHealth {
		if !isHealthy {
			healthy = false
			break
		}
	}

	log.Printf("Cluster Health - Node: %v, Pod: %v\n", cacheData.NodeHealth, cacheData.PodHealth)
	return healthy
}

func (h *K8sHealthChecker) StopInformers() {
	close(h.stopCh)
}

//func main() {
//	//kubeconfig := "/path/to/your/kubeconfig"
//	k8sUtil, err := NewDefaultClient()
//	if err != nil {
//		log.Fatalf("Error creating K8sUtil: %v", err)
//	}
//
//	// 启动 informers
//	k8sUtil.StartInformers()
//
//	// 使用 RegisterObserver 注册观察者
//	k8sUtil.RegisterObserver(func(data *ClusterHealthData) {
//		log.Printf("Cluster health updated: %+v", data)
//	})
//
//	if k8sUtil.IsClusterHealthy() {
//		log.Println("Kubernetes cluster is healthy")
//	} else {
//		log.Println("Kubernetes cluster is not healthy")
//	}
//
//	// Keep the main routine running to allow observer to work
//	select {}
//}
