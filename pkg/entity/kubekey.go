package entity

type KubekeyConf struct {
	ClusterName       string
	Hosts             []Host
	Etcds             []string
	ContronPlanes     []string
	Workers           []string
	NtpServers        []string
	Registry          Registry
	VIPServer         string
	KubePodsCIDR      string
	KubeServiceCIDR   string
	ContainerManager  string
	ProxyMode         string
	IPIPMode          string
	VxlanMode         string
	KKPath            string
	TaichuPackagePath string
	KubernetesVersion string
}

type KubekeyTemplate struct {
	ClusterName       string
	HostList          string
	EtcdList          string
	ControlPlaneList  string
	WorkerList        string
	NtpServerList     string
	Registry          string
	VIPServer         string
	KubePodsCIDR      string
	KubeServiceCIDR   string
	ContainerManager  string
	ProxyMode         string
	IPIPMode          string
	VxlanMode         string
	KKPath            string
	TaichuPackagePath string
	RegistryType      string
	RegistryUrI       string
	RegistryUser      string
	RegistryPassword  string
	RegistryKeyPath   string
	RegistryCertPath  string
	RegistrySkipTLS   bool
	RegistryPlainHttp bool
	RegistryNodeName  string
	KubernetesVersion string
	InsecureRegistry  string
}
