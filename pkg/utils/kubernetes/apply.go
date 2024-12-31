package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/ghodss/yaml"
	helm "github.com/mittwald/go-helm-client"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"sync"
	"time"
)

func (client *K8sClient) ApplyYAML(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		logger.GetLogger().Printf("Error reading file %s: %s", file, err)
		return err
	}

	var unstructuredData *unstructured.Unstructured
	if err := yaml.Unmarshal(data, &unstructuredData); err != nil {
		logger.GetLogger().Printf("Error unmarshalling YAML: %s", err)
		return err
	}
	obj := unstructuredData.Object

	apiVersion, kind := obj["apiVersion"].(string), obj["kind"].(string)
	if apiVersion == "" || kind == "" {
		err := fmt.Errorf("apiVersion or kind not found in YAML")
		logger.GetLogger().Println(err)
		return err
	}

	gvr, namespaced := getGVR(kind, apiVersion)
	if gvr == (schema.GroupVersionResource{}) {
		err := fmt.Errorf("unsupported kind: %s", kind)
		logger.GetLogger().Println(err)
		return err
	}

	resourceClient := client.DynamicClient.Resource(gvr)
	namespace := getNamespace(obj)

	if namespaced && namespace != "" {
		return applyInNamespace(resourceClient, namespace, unstructuredData)
	}

	return applyNonNamespaced(resourceClient, unstructuredData)
}

func (client *K8sClient) ApplyYAMLs(files []string) (*entity.ApplyResults, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var results []entity.SingleApplyResult
	var overallSuccess = true

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			var result entity.SingleApplyResult
			result.FileName = file
			if err := client.ApplyYAML(file); err != nil {
				result.Success = false
				result.Error = err.Error()
				mu.Lock()
				overallSuccess = false
				mu.Unlock()
			} else {
				result.Success = true
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return &entity.ApplyResults{
		OverallSuccess: overallSuccess,
		Results:        results,
	}, nil
}

func (client *K8sClient) DeployYAML(content string) error {
	var unstructuredData *unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(content), &unstructuredData); err != nil {
		logger.GetLogger().Printf("Error unmarshalling YAML: %s", err)
		return err
	}
	obj := unstructuredData.Object

	apiVersion, kind := obj["apiVersion"].(string), obj["kind"].(string)
	if apiVersion == "" || kind == "" {
		err := fmt.Errorf("apiVersion or kind not found in YAML")
		logger.GetLogger().Println(err)
		return err
	}

	gvr, namespaced := getGVR(kind, apiVersion)
	if gvr == (schema.GroupVersionResource{}) {
		err := fmt.Errorf("unsupported kind: %s", kind)
		logger.GetLogger().Println(err)
		return err
	}

	resourceClient := client.DynamicClient.Resource(gvr)
	namespace := getNamespace(obj)

	if namespaced && namespace != "" {
		return applyInNamespace(resourceClient, namespace, unstructuredData)
	}

	return applyNonNamespaced(resourceClient, unstructuredData)
}

func (client *K8sClient) DeployYAMLs(contents []string) (*entity.ApplyResults, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var results []entity.SingleApplyResult
	var overallSuccess = true

	for _, content := range contents {
		wg.Add(1)
		go func(content string) {
			defer wg.Done()
			var result entity.SingleApplyResult
			result.FileName = ""
			if err := client.DeployYAML(content); err != nil {
				result.Success = false
				result.Error = err.Error()
				mu.Lock()
				overallSuccess = false
				mu.Unlock()
			} else {
				result.Success = true
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(content)
	}

	wg.Wait()
	return &entity.ApplyResults{
		OverallSuccess: overallSuccess,
		Results:        results,
	}, nil
}

func (client *K8sClient) UpgradeDeployment(namespace, deploymentName, containerName, newImage string) error {
	deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting Deployment: %s", err.Error())
	}

	updated := false
	for i := range deployment.Spec.Template.Spec.Containers {
		if deployment.Spec.Template.Spec.Containers[i].Name == containerName {
			if deployment.Spec.Template.Spec.Containers[i].Image != newImage {
				deployment.Spec.Template.Spec.Containers[i].Image = newImage
				updated = true
			}
		}
	}

	if !updated {
		logger.GetLogger().Errorf("No update required for Deployment %s", deploymentName)
		return nil
	}

	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating Deployment: %s", err.Error())
	}

	if err = client.waitForRollout(namespace, deploymentName); err != nil {
		logger.GetLogger().Errorf("Rollout failed: %s", err.Error())
		return client.rollbackDeployment(namespace, deploymentName)
	}

	logger.GetLogger().Errorf("Deployment %s updated successfully", deploymentName)
	return nil
}

func (client *K8sClient) waitForRollout(namespace, deploymentName string) error {
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for rollout of Deployment %s", deploymentName)
		case <-ticker.C:
			deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error getting Deployment: %s", err.Error())
			}

			if deployment.Status.UpdatedReplicas == deployment.Status.Replicas &&
				deployment.Status.Replicas == deployment.Status.ReadyReplicas {
				return nil
			}
		}
	}
}

func (client *K8sClient) rollbackDeployment(namespace, deploymentName string) error {
	logger.GetLogger().Errorf("Rolling back Deployment %s", deploymentName)

	deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting Deployment: %s", err.Error())
	}

	replicaSets, err := client.Clientset.AppsV1().ReplicaSets(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentName), // Adjust this selector as needed
	})
	if err != nil {
		return fmt.Errorf("error listing ReplicaSets: %s", err.Error())
	}

	var currentRS *appsv1.ReplicaSet
	var previousRS *appsv1.ReplicaSet

	for _, rs := range replicaSets.Items {
		if labelsAreEqual(rs.Labels, deployment.Spec.Selector.MatchLabels) {
			currentRS = &rs
		}
	}

	for _, rs := range replicaSets.Items {
		if rs.Name != currentRS.Name {
			if previousRS == nil || compareReplicaSetRevision(rs, *previousRS) < 0 {
				previousRS = &rs
			}
		}
	}

	if previousRS == nil {
		return fmt.Errorf("no previous ReplicaSet found to roll back to")
	}

	deployment.Spec.Template.Spec.Containers[0].Image = previousRS.Spec.Template.Spec.Containers[0].Image
	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error rolling back Deployment: %s", err.Error())
	}

	logger.GetLogger().Errorf("Deployment %s rolled back to previous ReplicaSet %s", deploymentName, previousRS.Name)
	return nil
}

func labelsAreEqual(labels1, labels2 map[string]string) bool {
	if len(labels1) != len(labels2) {
		return false
	}
	for key, value := range labels1 {
		if labels2[key] != value {
			return false
		}
	}
	return true
}

func compareReplicaSetRevision(rs1, rs2 appsv1.ReplicaSet) int {
	return len(rs1.Name) - len(rs2.Name)
}

var (
	DefaultTimeout      = 10 * time.Minute
	DefaultPollInterval = 10 * time.Second
)

func (client *K8sClient) CheckPodStatuses(namespace string, timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			logger.GetLogger().Infof("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allPodsHealthy := true

			pods, err := client.Clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list pods: %v", err)
				return false
			}

			for _, pod := range pods.Items {
				status := pod.Status.Phase
				if status != corev1.PodRunning && status != corev1.PodSucceeded {
					logger.GetLogger().Errorf("Pod %s is in %s state\n", pod.Name, status)
					allPodsHealthy = false
				}
			}

			if allPodsHealthy {
				logger.GetLogger().Infof("All pods are in Running or Completed state. Application deployment is successful.")
				return true
			}

			logger.GetLogger().Infof("Some pods are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) CheckNodeStatuses(timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			fmt.Println("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allNodesHealthy := true

			// List all nodes
			nodes, err := client.Clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list nodes: %v", err)
				return false
			}

			for _, node := range nodes.Items {
				// Check the status of each node
				nodeStatus := "Unknown"
				for _, condition := range node.Status.Conditions {
					if condition.Type == corev1.NodeReady {
						nodeStatus = string(condition.Status)
						break
					}
				}
				if nodeStatus != string(v1.ConditionTrue) {
					logger.GetLogger().Errorf("Node %s is in %s state\n", node.Name, nodeStatus)
					allNodesHealthy = false
				}
			}

			if allNodesHealthy {
				logger.GetLogger().Infof("All nodes are in Ready state. Cluster is healthy.")
				return true
			}

			logger.GetLogger().Infof("Some nodes are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) CheckDeploymentStatuses(namespace string, timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			logger.GetLogger().Infof("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allDeploymentsHealthy := true

			deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list deployments: %v", err)
				return false
			}

			for _, deployment := range deployments.Items {
				if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
					logger.GetLogger().Errorf("Deployment %s has %d/%d replicas ready", deployment.Name, deployment.Status.ReadyReplicas, deployment.Status.Replicas)
					allDeploymentsHealthy = false
				}
			}

			if allDeploymentsHealthy {
				logger.GetLogger().Info("All deployments are healthy.")
				return true
			}

			logger.GetLogger().Info("Some deployments are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) CheckStatefulSetStatuses(namespace string, timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			logger.GetLogger().Info("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allStatefulSetsHealthy := true

			statefulSets, err := client.Clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list statefulsets: %v", err)
				return false
			}

			for _, statefulSet := range statefulSets.Items {
				if statefulSet.Status.ReadyReplicas != statefulSet.Status.Replicas {
					logger.GetLogger().Errorf("StatefulSet %s has %d/%d replicas ready", statefulSet.Name, statefulSet.Status.ReadyReplicas, statefulSet.Status.Replicas)
					allStatefulSetsHealthy = false
				}
			}

			if allStatefulSetsHealthy {
				logger.GetLogger().Info("All statefulsets are healthy.")
				return true
			}

			logger.GetLogger().Info("Some statefulsets are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) CheckDaemonSetStatuses(namespace string, timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			logger.GetLogger().Info("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allDaemonSetsHealthy := true

			daemonSets, err := client.Clientset.AppsV1().DaemonSets(namespace).List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list daemonsets: %v", err)
				return false
			}

			for _, daemonSet := range daemonSets.Items {
				desiredNumberScheduled := daemonSet.Status.DesiredNumberScheduled
				numberReady := daemonSet.Status.NumberReady
				if desiredNumberScheduled != numberReady {
					logger.GetLogger().Errorf("DaemonSet %s has %d/%d pods scheduled", daemonSet.Name, numberReady, desiredNumberScheduled)
					allDaemonSetsHealthy = false
				}
			}

			if allDaemonSetsHealthy {
				logger.GetLogger().Info("All daemonsets are healthy.")
				return true
			}

			logger.GetLogger().Info("Some daemonsets are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) CheckPVCStatuses(namespace string, timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			logger.GetLogger().Info("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allPVCsHealthy := true

			pvcs, err := client.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list pvc: %v", err)
				return false
			}

			for _, pvc := range pvcs.Items {
				if pvc.Status.Phase != corev1.ClaimBound {
					logger.GetLogger().Errorf("PVC %s is in %s state", pvc.Name, pvc.Status.Phase)
					allPVCsHealthy = false
				}
			}

			if allPVCsHealthy {
				logger.GetLogger().Info("All PVCs are in Bound state.")
				return true
			}

			logger.GetLogger().Info("Some PVCs are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) CheckPVStatuses(timeout time.Duration) bool {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			logger.GetLogger().Info("Timeout reached. Exiting.")
			return false
		case <-ticker.C:
			allPVsHealthy := true

			pvs, err := client.Clientset.CoreV1().PersistentVolumes().List(context.TODO(), v1.ListOptions{})
			if err != nil {
				logger.GetLogger().Errorf("Failed to list pv: %v", err)
				return false
			}

			for _, pv := range pvs.Items {
				if pv.Status.Phase != corev1.VolumeBound {
					logger.GetLogger().Errorf("PV %s is in %s state", pv.Name, pv.Status.Phase)
					allPVsHealthy = false
				}
			}

			if allPVsHealthy {
				logger.GetLogger().Info("All PVs are in Bound state.")
				return true
			}

			logger.GetLogger().Info("Some PVs are not in the expected state. Continuing to check...")
		}
	}
}

func (client *K8sClient) GetServiceIPAndPort(namespace, serviceName string) (string, []corev1.ServicePort, error) {
	service, err := client.Clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, v1.GetOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Extract IP and Ports
	serviceIP := service.Spec.ClusterIP
	ports := service.Spec.Ports

	return serviceIP, ports, nil
}

func (client *K8sClient) GetSecretData(namespace, secretName string) (map[string]string, error) {
	secret, err := client.Clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Extract key-value pairs from the secret
	secretData := make(map[string]string)
	for key, value := range secret.Data {
		secretData[key] = string(value) // Secret data is base64 encoded, decode if needed
	}

	return secretData, nil
}

func (client *K8sClient) GetSecretDecodeData(namespace, secretName string) (map[string]string, error) {
	secret, err := client.Clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		logger.GetLogger().Errorf("Error getting secret %s: %v", secretName, err)
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	secretData := make(map[string]string)
	for key, value := range secret.Data {
		decodedValue, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			logger.GetLogger().Errorf("Error decoding base64 for key %s: %v", key, err)
			return nil, fmt.Errorf("failed to decode base64 for key %s: %w", key, err)
		}
		secretData[key] = string(decodedValue)
	}

	return secretData, nil
}

func (client *K8sClient) GetConfigMapData(namespace, configMapName string) (map[string]string, error) {
	configMap, err := client.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	// Extract key-value pairs from the configmap
	configMapData := make(map[string]string)
	for key, value := range configMap.Data {
		configMapData[key] = value
	}

	return configMapData, nil
}

func (client *K8sClient) GetIngressInfo(namespace, ingressName string) ([]string, error) {
	ingress, err := client.Clientset.NetworkingV1().Ingresses(namespace).Get(context.TODO(), ingressName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	var hostInfo []string
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			hostInfo = append(hostInfo, fmt.Sprintf("Host: %s, Path: %s", rule.Host, path.Path))
		}
	}

	return hostInfo, nil
}

func (client *K8sClient) ListNamespaces() ([]string, error) {
	namespaces, err := client.Clientset.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var nsList []string
	for _, ns := range namespaces.Items {
		nsList = append(nsList, ns.Name)
	}

	return nsList, nil
}

func (client *K8sClient) GetPodInfo(namespace, podName string) (*corev1.Pod, error) {
	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return pod, nil
}

func (client *K8sClient) GetDeployInfo(namespace, deployName string) (*appsv1.Deployment, error) {
	deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deployName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return deployment, nil
}

func (client *K8sClient) GetStatefulSetInfo(namespace, statefulSetName string) (*appsv1.StatefulSet, error) {
	statefulSet, err := client.Clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset: %w", err)
	}

	return statefulSet, nil
}

func (client *K8sClient) GetDaemonSetInfo(namespace, daemonSetName string) (*appsv1.DaemonSet, error) {
	daemonSet, err := client.Clientset.AppsV1().DaemonSets(namespace).Get(context.TODO(), daemonSetName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonset: %w", err)
	}

	return daemonSet, nil
}

func (client *K8sClient) GetSvcInfo(namespace, svcName string) (*corev1.Service, error) {
	service, err := client.Clientset.CoreV1().Services(namespace).Get(context.TODO(), svcName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return service, nil
}

func (client *K8sClient) GetServiceAccountInfo(namespace, serviceAccountName string) ([]string, error) {
	serviceAccount, err := client.Clientset.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), serviceAccountName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service account: %w", err)
	}

	// Extract secret names from the ServiceAccount object
	var secretNames []string
	for _, secretRef := range serviceAccount.Secrets {
		secretNames = append(secretNames, secretRef.Name)
	}

	return secretNames, nil
}

func (client *K8sClient) GetRolePermissions(namespace, roleName string) ([]string, error) {
	role, err := client.Clientset.RbacV1().Roles(namespace).Get(context.TODO(), roleName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	var permissions []string
	for _, rule := range role.Rules {
		permissions = append(permissions, fmt.Sprintf("APIGroups: %v, Resources: %v, Verbs: %v", rule.APIGroups, rule.Resources, rule.Verbs))
	}

	return permissions, nil
}

func (client *K8sClient) GetRoleBindingSubjects(namespace, roleBindingName string) ([]string, error) {
	roleBinding, err := client.Clientset.RbacV1().RoleBindings(namespace).Get(context.TODO(), roleBindingName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get rolebinding: %w", err)
	}

	var subjects []string
	for _, subject := range roleBinding.Subjects {
		subjects = append(subjects, fmt.Sprintf("Kind: %s, Name: %s, Namespace: %s", subject.Kind, subject.Name, subject.Namespace))
	}

	return subjects, nil
}

func (client *K8sClient) GetClusterRoleInfo(clusterRoleName string) (*rbacv1.ClusterRole, error) {
	clusterRole, err := client.Clientset.RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterRole: %w", err)
	}

	return clusterRole, nil
}

func (client *K8sClient) GetClusterRoleBindingInfo(clusterRoleBindingName string) (*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBinding, err := client.Clientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterRoleBinding: %w", err)
	}

	return clusterRoleBinding, nil
}

func (client *K8sClient) GetCustomResource(namespace, resourceName, group, version, resourceKind string) (interface{}, error) {
	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceKind,
	}

	resource, err := client.DynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), resourceName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get custom resource: %w", err)
	}

	return resource, nil
}

func (client *K8sClient) GetClusterCustomResource(resourceName, group, version, resourceKind string) (interface{}, error) {
	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceKind,
	}

	resource := client.DynamicClient.Resource(gvr)

	customResource, err := resource.Get(context.TODO(), resourceName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster custom resource: %w", err)
	}

	return customResource, nil
}

func (client *K8sClient) GetCronJobInfo(namespace, cronJobName string) (*batchv1.CronJob, error) {
	cronJob, err := client.Clientset.BatchV1().CronJobs(namespace).Get(context.TODO(), cronJobName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}

	return cronJob, nil
}

func (client *K8sClient) GetHPAInfo(namespace, hpaName string) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	hpa, err := client.Clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Get(context.TODO(), hpaName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get HPA: %w", err)
	}

	return hpa, nil
}

func (client *K8sClient) GetPDBInfo(namespace, pdbName string) (*policyv1.PodDisruptionBudget, error) {
	pdb, err := client.Clientset.PolicyV1().PodDisruptionBudgets(namespace).Get(context.TODO(), pdbName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PodDisruptionBudget: %w", err)
	}

	return pdb, nil
}

func (client *K8sClient) GetEvents(namespace string) ([]corev1.Event, error) {
	events, err := client.Clientset.CoreV1().Events(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	return events.Items, nil
}

func (client *K8sClient) GetConfigMap(namespace, configMapName string) (*corev1.ConfigMap, error) {
	configMap, err := client.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap: %w", err)
	}
	return configMap, nil
}

func (client *K8sClient) GetSecret(namespace, secretName string) (*corev1.Secret, error) {
	secret, err := client.Clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Secret: %w", err)
	}
	return secret, nil
}

func (client *K8sClient) GetPodResources(namespace, podName string) (*corev1.Pod, error) {
	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Pod: %w", err)
	}
	return pod, nil
}

func (client *K8sClient) GetNodeInfo(nodeName string) (*corev1.Node, error) {
	node, err := client.Clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Node: %w", err)
	}
	return node, nil
}

func (client *K8sClient) GetNetworkPolicy(namespace, networkPolicyName string) (*networkingv1.NetworkPolicy, error) {
	networkPolicy, err := client.Clientset.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), networkPolicyName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get NetworkPolicy: %w", err)
	}
	return networkPolicy, nil
}

func (client *K8sClient) GetCRDInfo(crdName string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd, err := client.CRDClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crdName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CRD: %w", err)
	}
	return crd, nil
}

func (client *K8sClient) AddOrUpdateChartRepo(helmRepo entity.HelmRepository) error {
	var repoEntry repo.Entry
	repoEntry.Name = helmRepo.Name
	repoEntry.URL = helmRepo.Url
	repoEntry.Username = helmRepo.Username
	repoEntry.Password = helmRepo.Password
	repoEntry.CertFile = helmRepo.CertFile
	repoEntry.CAFile = helmRepo.CAFile
	repoEntry.KeyFile = helmRepo.KeyFile
	repoEntry.InsecureSkipTLSverify = helmRepo.InsecureSkipTlsVerify
	if err := client.HelmClient.AddOrUpdateChartRepo(repoEntry); err != nil {
		logger.GetLogger().Errorf("Faile to add or update helm repository to cluster: %s", err.Error())
		return err
	}
	return nil
}

func (client *K8sClient) InstallOrUpgradeChart(info entity.HelmChartInfo) (*release.Release, error) {
	chartSpec := helm.ChartSpec{
		ReleaseName:     info.ReleaseName,
		ChartName:       info.ChartName,
		Namespace:       info.Namespace,
		ValuesYaml:      info.ValuesYaml,
		CreateNamespace: info.CreateNamespace,
	}
	release1, err := client.HelmClient.InstallOrUpgradeChart(context.TODO(), &chartSpec, &helm.GenericHelmOptions{})
	if err != nil {
		logger.GetLogger().Errorf("Faile to install or update chart release: %s", err.Error())
		return nil, err
	}
	return release1, nil
}

func (client *K8sClient) ListDeployedReleases() ([]*release.Release, error) {
	releases, err := client.HelmClient.ListDeployedReleases()
	if err != nil {
		logger.GetLogger().Errorf("Faile to list deployed helm chart release: %s", err.Error())
		return nil, err
	}
	return releases, nil
}

func (client *K8sClient) UninstallRelease(info entity.HelmChartInfo) error {
	chartSpec := helm.ChartSpec{
		ReleaseName:     info.ReleaseName,
		ChartName:       info.ChartName,
		Namespace:       info.Namespace,
		ValuesYaml:      info.ValuesYaml,
		CreateNamespace: info.CreateNamespace,
	}
	err := client.HelmClient.UninstallRelease(&chartSpec)
	if err != nil {
		logger.GetLogger().Errorf("Faile to uninstall deployed helm chart release[%s]: %s", info.ReleaseName, err.Error())
		return err
	}
	return nil
}

func (client *K8sClient) GetHelmCharts(repoURL string) (map[string][]entity.Chart, error) {
	resp, err := client.HttpClient.Get(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get index.yaml: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var index entity.RepoIndex
	if err := yaml.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index.yaml: %w", err)
	}

	return index.Entries, nil
}
