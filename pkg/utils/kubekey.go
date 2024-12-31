package utils

import (
	"bytes"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"log"
	"path/filepath"
	"strings"
	"text/template"
)

type KubekeyClient struct {
	KubekeyConf entity.KubekeyConf
	OSClient    OSClient
}

func NewKubekeyClient(kubekeyConf entity.KubekeyConf, osClient OSClient) *KubekeyClient {
	return &KubekeyClient{
		KubekeyConf: kubekeyConf,
		OSClient:    osClient,
	}
}

func (client *KubekeyClient) ParseToTemplate() *entity.KubekeyTemplate {
	template := &entity.KubekeyTemplate{}
	for _, host := range client.KubekeyConf.Hosts {
		template.HostList += fmt.Sprintf("- {name: %s, address: %s, internalAddress: %s, port: %d, user: %s, password: %s}\n  ", host.Name, host.Address, host.InternalAddress, host.Port, host.User, host.Password)
		if host.Registry != nil && host.Registry.InsecureRegistries != nil {
			for _, ir := range host.Registry.InsecureRegistries {
				template.InsecureRegistry += fmt.Sprintf("%s", ir)
				template.InsecureRegistry += ","
			}
			template.InsecureRegistry = template.InsecureRegistry[:len(template.InsecureRegistry)-1]
		}
	}
	template.HostList = strings.TrimSpace(template.HostList)

	for _, cp := range client.KubekeyConf.Etcds {
		template.EtcdList += fmt.Sprintf("- %s\n    ", cp)
	}
	template.EtcdList = strings.TrimSpace(template.EtcdList)

	for _, cp := range client.KubekeyConf.ContronPlanes {
		template.ControlPlaneList += fmt.Sprintf("- %s\n    ", cp)
	}
	template.ControlPlaneList = strings.TrimSpace(template.ControlPlaneList)

	for _, cp := range client.KubekeyConf.Workers {
		template.WorkerList += fmt.Sprintf("- %s\n    ", cp)
	}
	template.WorkerList = strings.TrimSpace(template.WorkerList)

	for _, ns := range client.KubekeyConf.NtpServers {
		template.NtpServerList += fmt.Sprintf("- %s\n      ", ns)
	}

	template.NtpServerList = strings.TrimSpace(template.NtpServerList)
	template.Registry += fmt.Sprintf("- %s", client.KubekeyConf.Registry.NodeName)
	template.RegistryType = client.KubekeyConf.Registry.Type
	template.RegistryUrI = client.KubekeyConf.Registry.Url
	template.RegistryUser = client.KubekeyConf.Registry.User
	template.RegistryPassword = client.KubekeyConf.Registry.Password
	template.RegistryCertPath = client.KubekeyConf.Registry.CertPath
	template.RegistryKeyPath = client.KubekeyConf.Registry.KeyPath
	template.RegistrySkipTLS = client.KubekeyConf.Registry.SkipTLS
	template.RegistryPlainHttp = client.KubekeyConf.Registry.PlainHttp
	template.RegistryNodeName = client.KubekeyConf.Registry.NodeName
	template.ProxyMode = client.KubekeyConf.ProxyMode
	template.IPIPMode = client.KubekeyConf.IPIPMode
	template.VxlanMode = client.KubekeyConf.VxlanMode
	template.ContainerManager = client.KubekeyConf.ContainerManager
	template.ClusterName = client.KubekeyConf.ClusterName
	template.KubernetesVersion = client.KubekeyConf.KubernetesVersion
	template.KubeServiceCIDR = client.KubekeyConf.KubeServiceCIDR
	template.KubePodsCIDR = client.KubekeyConf.KubePodsCIDR
	template.KKPath = client.KubekeyConf.KKPath
	template.TaichuPackagePath = client.KubekeyConf.TaichuPackagePath
	template.VIPServer = client.KubekeyConf.VIPServer
	return template
}

func (client *KubekeyClient) GenerateConfig() error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	templateText := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: {{ .ClusterName }}
spec:
  hosts:
  {{ .HostList }}
  roleGroups:
    etcd:
    {{ .EtcdList }}
    control-plane: 
    {{ .ControlPlaneList }}
    worker:
    {{ .WorkerList }}
    registry:
    {{ .Registry }}
  controlPlaneEndpoint:
    internalLoadbalancer: haproxy
    domain: lb.cars.local
    address: ""
    port: 6443
  system:
    ntpServers:
      {{ .NtpServerList }}
    timezone: "Asia/Shanghai"
  kubernetes:
    version: {{ .KubernetesVersion }}
    clusterName: cluster.local
    autoRenewCerts: true
    containerManager: {{ .ContainerManager }}
    apiserverCertExtraSans:  
      - lb.cars.local
    proxyMode: {{ .ProxyMode }}
  etcd:
    type: kubekey
  network:
    plugin: calico
    calico:
      ipipMode: {{ .IPIPMode }}
      vxlanMode: {{ .VxlanMode }}
    kubePodsCIDR: {{ .KubePodsCIDR }}
    kubeServiceCIDR: {{ .KubeServiceCIDR }}
    multusCNI:
      enabled: false
  registry:
    {{- if .RegistryType }}
    type: {{ .RegistryType }}
    {{- end }}
    auths:
      "{{ .RegistryUrI }}":
        username: {{ .RegistryUser }}
        password: {{ .RegistryPassword }}
        skipTLSVerify: {{ .RegistrySkipTLS }}
        plainHTTP: {{ .RegistryPlainHttp }}
    privateRegistry: "{{ .RegistryUrI }}"
    namespaceOverride: "carsio"
    registryMirrors: []
    insecureRegistries: ["{{ .InsecureRegistry }}"]
  addons: []
`
	kubekeyTemplate := client.ParseToTemplate()
	tmpl, err := template.New("config-sample.yaml").Parse(templateText)
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate template object: %s", err.Error())
		return err
	}
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, kubekeyTemplate)
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate template: %s", err.Error())
		return err
	}
	err = client.OSClient.SSExecutor.MkDirALL(configPath, func(s string) {
		logger.GetLogger().Infof(s)
	})
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate dir %s: %s", configPath, err.Error())
		return err
	}
	command := fmt.Sprintf("bash -c \"echo '%s' > %s\"", rendered.String(), path)
	if client.OSClient.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.OSClient.SSExecutor.Host.Password)
	}
	err = client.OSClient.SSExecutor.ExecuteCommandWithoutReturn(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate kubekey config: %s", err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) GenerateConfigWithVIP() error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	templateText := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: {{ .ClusterName }}
spec:
  hosts:
  {{ .HostList }}
  roleGroups:
    etcd:
    {{ .EtcdList }}
    control-plane: 
    {{ .ControlPlaneList }}
    worker:
    {{ .WorkerList }}
    registry:
    {{ .Registry }}
  controlPlaneEndpoint:
    domain: lb.cars.local
    address: "{{ .VIPServer }}"
    port: 6443
  system:
    ntpServers:
      {{ .NtpServerList }}
    timezone: "Asia/Shanghai"
  kubernetes:
    version: {{ .KubernetesVersion }}
    clusterName: cluster.local
    autoRenewCerts: true
    containerManager: {{ .ContainerManager }}
    apiserverCertExtraSans:  
      - lb.cars.local
    proxyMode: {{ .ProxyMode }}
  etcd:
    type: kubekey
  network:
    plugin: calico
    calico:
      ipipMode: {{ .IPIPMode }}
      vxlanMode: {{ .VxlanMode }}
    kubePodsCIDR: {{ .KubePodsCIDR }}
    kubeServiceCIDR: {{ .KubeServiceCIDR }}
    multusCNI:
      enabled: false
  registry:
    {{- if .RegistryType }}
    type: {{ .RegistryType }}
    {{- end }}
    auths:
      "{{ .RegistryUrI }}":
        username: {{ .RegistryUser }}
        password: {{ .RegistryPassword }}
        skipTLSVerify: {{ .RegistrySkipTLS }}
        plainHTTP: {{ .RegistryPlainHttp }}
    privateRegistry: "{{ .RegistryUrI }}"
    namespaceOverride: "carsio"
    registryMirrors: []
    insecureRegistries: ["{{ .InsecureRegistry }}"]
  addons: []
`
	kubekeyTemplate := client.ParseToTemplate()
	tmpl, err := template.New("config-sample.yaml").Parse(templateText)
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate template object: %s", err.Error())
		return err
	}
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, kubekeyTemplate)
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate template: %s", err.Error())
		return err
	}
	err = client.OSClient.SSExecutor.MkDirALL(configPath, func(s string) {
		log.Println(s)
	})
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate dir %s: %s", configPath, err.Error())
		return err
	}
	//command := fmt.Sprintf("echo '%s' > %s", rendered.String(), path)
	command := fmt.Sprintf("bash -c \"echo '%s' > %s\"", rendered.String(), path)
	if client.OSClient.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.OSClient.SSExecutor.Host.Password)
	}
	err = client.OSClient.SSExecutor.ExecuteCommandWithoutReturn(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate kubekey config: %s", err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) CreateCluster(logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk create cluster -f %s -a %s --with-packages --yes", path, client.KubekeyConf.TaichuPackagePath)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create cluster %s: %s", client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) DeleteCluster(logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk delete cluster -f %s --yes", path)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to delete cluster %s: %s", client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) AddNode(logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk add nodes -f %s --yes", path)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to add node to cluster %s: %s", client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) DeleteNode(nodeName string, logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk delete node %s -f %s", nodeName, path)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to delete node %s from cluster %s: %s", nodeName, client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) CheckCertExpirtation(logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk certs check-expirtation -f %s", path)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to check cert expiration %s: %s", client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) RenewCert(logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk certs renew -f %s", path)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to renew cert %s: %s", client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}

func (client *KubekeyClient) UpgradeCluster(logChan chan LogEntry) error {
	dirPath := filepath.Dir(client.KubekeyConf.KKPath)
	configPath := filepath.Join(dirPath, client.KubekeyConf.ClusterName)
	configPath = filepath.ToSlash(configPath)
	path := filepath.Join(configPath, "config-sample.yaml")
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("kk upgrade -f %s", path)
	err := client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Errorf("Failed to upgrade %s: %s", client.KubekeyConf.ClusterName, err.Error())
		return err
	}
	return nil
}
