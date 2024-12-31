package kubeadm

import "time"

type ClusterConfiguration struct {
	Etcd                 Etcd              `yaml:"etcd" json:"etcd"`
	DNS                  DNS               `yaml:"dns" json:"dns"`
	ImageRepository      string            `yaml:"imageRepository" json:"imageRepository"`
	KubernetesVersion    string            `yaml:"kubernetesVersion" json:"kubernetesVersion"`
	CertificatesDir      string            `yaml:"certificatesDir" json:"certificatesDir"`
	ClusterName          string            `yaml:"clusterName" json:"clusterName"`
	ControlPlaneEndpoint string            `yaml:"controlPlaneEndpoint" json:"controlPlaneEndpoint"`
	Networking           Networking        `yaml:"networking" json:"networking"`
	APIServer            APIServer         `yaml:"apiServer" json:"apiServer"`
	CertSANs             []string          `yaml:"certSANs" json:"certSANs"`
	ControllerManager    ControllerManager `yaml:"controllerManager" json:"controllerManager"`
	Scheduler            Scheduler         `yaml:"scheduler" json:"scheduler"`
}

type Etcd struct {
	External External `yaml:"external" json:"external"`
}

type External struct {
	Endpoints []string `yaml:"endpoints" json:"endpoints"`
	CaFile    string   `yaml:"caFile" json:"caFile"`
	CertFile  string   `yaml:"certFile" json:"certFile"`
	KeyFile   string   `yaml:"keyFile" json:"keyFile"`
}

type DNS struct {
	Type            string `yaml:"type" json:"type"`
	ImageRepository string `yaml:"imageRepository" json:"imageRepository"`
	ImageTag        string `yaml:"imageTag" json:"imageTag"`
}

type Networking struct {
	DNSDomain     string `yaml:"dnsDomain" json:"dnsDomain"`
	PodSubnet     string `yaml:"podSubnet" json:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet" json:"serviceSubnet"`
}

type APIServer struct {
	ExtraArgs    ExtraArgs     `yaml:"extraArgs" json:"extraArgs"`
	CertSANs     []string      `yaml:"certSANs" json:"certSANs"`
	ExtraVolumes []ExtraVolume `yaml:"extraVolumes" json:"extraVolumes"`
}

type ExtraArgs struct {
	AuditLogPath           string `yaml:"audit-log-path" json:"audit-log-path"`
	AuditLogMaxAge         string `yaml:"audit-log-maxage" json:"audit-log-maxage"`
	AuditLogMaxBackup      string `yaml:"audit-log-maxbackup" json:"audit-log-maxbackup"`
	AuditLogMaxSize        string `yaml:"audit-log-maxsize" json:"audit-log-maxsize"`
	AuditPolicyFile        string `yaml:"audit-policy-file" json:"audit-policy-file"`
	AuthorizationMode      string `yaml:"authorization-mode" json:"authorization-mode"`
	EnableAdmissionPlugins string `yaml:"enable-admission-plugins" json:"enable-admission-plugins"`
	AnonymousAuth          string `yaml:"anonymous-auth" json:"anonymous-auth"`
	BindAddress            string `yaml:"bind-address" json:"bind-address"`
	InsecureBindAddress    string `yaml:"insecure-bind-address" json:"insecure-bind-address"`
	InsecurePort           string `yaml:"insecure-port" json:"insecure-port"`
	TlsCertFile            string `yaml:"tls-cert-file" json:"tls-cert-file"`
	TlsPrivateKeyFile      string `yaml:"tls-private-key-file" json:"tls-private-key-file"`
	LogLevel               string `yaml:"log-level" json:"log-level"`
	FeatureGates           string `yaml:"feature-gates" json:"feature-gates"`
}

type ControllerManager struct {
	ExtraArgs    ExtraArgsCM   `yaml:"extraArgs" json:"extraArgs"`
	ExtraVolumes []ExtraVolume `yaml:"extraVolumes" json:"extraVolumes"`
}

type ExtraArgsCM struct {
	NodeCIDRMaskSize       string        `yaml:"node-cidr-mask-size" json:"node-cidr-mask-size"`
	BindAddress            string        `yaml:"bind-address" json:"bind-address"`
	ClusterSigningDuration time.Duration `yaml:"cluster-signing-duration" json:"cluster-signing-duration"`
	FeatureGates           string        `yaml:"feature-gates" json:"feature-gates"`
}

type ExtraVolume struct {
	Name      string `yaml:"name" json:"name"`
	HostPath  string `yaml:"hostPath" json:"hostPath"`
	MountPath string `yaml:"mountPath" json:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly" json:"readOnly"`
	PathType  string `yaml:"pathType" json:"pathType"`
}

type Scheduler struct {
	ExtraArgs ExtraArgsScheduler `yaml:"extraArgs" json:"extraArgs"`
}

type ExtraArgsScheduler struct {
	BindAddress  string `yaml:"bind-address" json:"bind-address"`
	FeatureGates string `yaml:"feature-gates" json:"feature-gates"`
}
