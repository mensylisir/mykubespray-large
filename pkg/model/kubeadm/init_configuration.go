package kubeadm

type InitConfiguration struct {
	LocalAPIEndpoint LocalAPIEndpoint `yaml:"localAPIEndpoint" json:"localAPIEndpoint"`
	NodeRegistration NodeRegistration `yaml:"nodeRegistration" json:"nodeRegistration"`
}

type LocalAPIEndpoint struct {
	AdvertiseAddress string `yaml:"advertiseAddress" json:"advertiseAddress"`
	BindPort         int    `yaml:"bindPort" json:"bindPort"`
}

type NodeRegistration struct {
	CRISocket        string           `yaml:"criSocket" json:"criSocket"`
	KubeletExtraArgs KubeletExtraArgs `yaml:"kubeletExtraArgs" json:"kubeletExtraArgs"`
}

type KubeletExtraArgs struct {
	CgroupDriver string `yaml:"cgroup-driver" json:"cgroup-driver"`
}
