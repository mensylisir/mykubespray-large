package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

var gvrMapping = map[string]schema.GroupVersionResource{
	"Pod":                {Group: "", Version: "v1", Resource: "pods"},
	"Service":            {Group: "", Version: "v1", Resource: "services"},
	"Deployment":         {Group: "apps", Version: "v1", Resource: "deployments"},
	"StatefulSet":        {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"DaemonSet":          {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"ReplicaSet":         {Group: "apps", Version: "v1", Resource: "replicasets"},
	"Job":                {Group: "batch", Version: "v1", Resource: "jobs"},
	"CronJob":            {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"ConfigMap":          {Group: "", Version: "v1", Resource: "configmaps"},
	"Secret":             {Group: "", Version: "v1", Resource: "secrets"},
	"Namespace":          {Group: "", Version: "v1", Resource: "namespaces"},
	"Ingress":            {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"NetworkPolicy":      {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	"ResourceQuota":      {Group: "", Version: "v1", Resource: "resourcequotas"},
	"LimitRange":         {Group: "", Version: "v1", Resource: "limitranges"},
	"Role":               {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	"ClusterRole":        {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	"RoleBinding":        {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	"ClusterRoleBinding": {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	"ServiceAccount":     {Group: "", Version: "v1", Resource: "serviceaccounts"},
}

func getGVR(kind, apiVersion string) (schema.GroupVersionResource, bool) {
	if gvr, ok := gvrMapping[kind]; ok {
		return gvr, true
	}
	parts := strings.SplitN(apiVersion, "/", 2)
	if len(parts) == 2 {
		group := parts[0]
		version := parts[1]
		pluralKind := toPlural(kind)
		return schema.GroupVersionResource{Group: group, Version: version, Resource: pluralKind}, true
	}

	return schema.GroupVersionResource{}, false
}

func toPlural(kind string) string {
	switch kind {
	case "Pod":
		return "pods"
	case "Service":
		return "services"
	case "Deployment":
		return "deployments"
	case "StatefulSet":
		return "statefulsets"
	case "DaemonSet":
		return "daemonsets"
	case "ReplicaSet":
		return "replicasets"
	case "Job":
		return "jobs"
	case "CronJob":
		return "cronjobs"
	case "ConfigMap":
		return "configmaps"
	case "Secret":
		return "secrets"
	case "Namespace":
		return "namespaces"
	case "Ingress":
		return "ingresses"
	case "NetworkPolicy":
		return "networkpolicies"
	case "ResourceQuota":
		return "resourcequotas"
	case "LimitRange":
		return "limitranges"
	case "Role":
		return "roles"
	case "ClusterRole":
		return "clusterroles"
	case "RoleBinding":
		return "rolebindings"
	case "ClusterRoleBinding":
		return "clusterrolebindings"
	case "ServiceAccount":
		return "serviceaccounts"
	default:
		return ""
	}
}
