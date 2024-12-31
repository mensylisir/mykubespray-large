package kubeadm

type KubeProxyConfiguration struct {
	ClusterCIDR string   `yaml:"clusterCIDR" json:"clusterCIDR"`
	IPTables    IPTables `yaml:"iptables" json:"iptables"`
	Mode        string   `yaml:"mode" json:"mode"`
}

type IPTables struct {
	MasqueradeAll bool   `yaml:"masqueradeAll" json:"masqueradeAll"`
	MasqueradeBit int    `yaml:"masqueradeBit" json:"masqueradeBit"`
	MinSyncPeriod string `yaml:"minSyncPeriod" json:"minSyncPeriod"`
	SyncPeriod    string `yaml:"syncPeriod" json:"syncPeriod"`
}
