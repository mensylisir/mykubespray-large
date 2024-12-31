package kubeadm

type KubeletConfiguration struct {
	ClusterDNS                       []string                `yaml:"clusterDNS" json:"clusterDNS"`
	ClusterDomain                    string                  `yaml:"clusterDomain" json:"clusterDomain"`
	ContainerLogMaxFiles             int                     `yaml:"containerLogMaxFiles" json:"containerLogMaxFiles"`
	ContainerLogMaxSize              string                  `yaml:"containerLogMaxSize" json:"containerLogMaxSize"`
	EvictionHard                     EvictionHard            `yaml:"evictionHard" json:"evictionHard"`
	EvictionMaxPodGracePeriod        int                     `yaml:"evictionMaxPodGracePeriod" json:"evictionMaxPodGracePeriod"`
	EvictionPressureTransitionPeriod string                  `yaml:"evictionPressureTransitionPeriod" json:"evictionPressureTransitionPeriod"`
	EvictionSoft                     EvictionSoft            `yaml:"evictionSoft" json:"evictionSoft"`
	EvictionSoftGracePeriod          EvictionSoftGracePeriod `yaml:"evictionSoftGracePeriod" json:"evictionSoftGracePeriod"`
	FeatureGates                     FeatureGates            `yaml:"featureGates" json:"featureGates"`
	KubeReserved                     KubeReserved            `yaml:"kubeReserved" json:"kubeReserved"`
	MaxPods                          int                     `yaml:"maxPods" json:"maxPods"`
	PodPidsLimit                     int                     `yaml:"podPidsLimit" json:"podPidsLimit"`
	RotateCertificates               bool                    `yaml:"rotateCertificates" json:"rotateCertificates"`
	SystemReserved                   SystemReserved          `yaml:"systemReserved" json:"systemReserved"`
}

type EvictionHard struct {
	MemoryAvailable string `yaml:"memory.available" json:"memory.available"`
	PIDAvailable    string `yaml:"pid.available" json:"pid.available"`
}

type EvictionSoft struct {
	MemoryAvailable string `yaml:"memory.available" json:"memory.available"`
}

type EvictionSoftGracePeriod struct {
	MemoryAvailable string `yaml:"memory.available" json:"memory.available"`
}

type FeatureGates struct {
	CSIStorageCapacity             bool `yaml:"CSIStorageCapacity" json:"CSIStorageCapacity"`
	ExpandCSIVolumes               bool `yaml:"ExpandCSIVolumes" json:"ExpandCSIVolumes"`
	RotateKubeletServerCertificate bool `yaml:"RotateKubeletServerCertificate" json:"RotateKubeletServerCertificate"`
}

type KubeReserved struct {
	CPU    string `yaml:"cpu" json:"cpu"`
	Memory string `yaml:"memory" json:"memory"`
}

type SystemReserved struct {
	CPU    string `yaml:"cpu" json:"cpu"`
	Memory string `yaml:"memory" json:"memory"`
}
