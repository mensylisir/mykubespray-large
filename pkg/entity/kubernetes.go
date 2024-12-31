package entity

type K8sConfig struct {
	Kubeconfig     string
	KubeconfigPath string
	ApiServer      string
	Token          string
	Cacert         string
}

type KubernetesFilesConf struct {
	K8sConfig
	Files []string
}

type SingleApplyResult struct {
	FileName string
	Success  bool
	Error    string
}

type ApplyResults struct {
	OverallSuccess bool
	Results        []SingleApplyResult
}
