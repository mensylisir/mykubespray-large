package entity

type HelmRepository struct {
	K8sConfig
	Name                  string `json:"name" gorm:"not null;unique"`
	Url                   string `json:"url" gorm:"not null;unique"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	IsActive              bool   `json:"is_active" gorm:"type:boolean;default:true"`
	CertFile              string `json:"cert_file"`
	KeyFile               string `json:"key_file"`
	CAFile                string `json:"ca_file"`
	InsecureSkipTlsVerify bool   `json:"insecure_skip_tls_verify"`
}

type HelmChartInfo struct {
	K8sConfig
	Namespace       string `json:"namespace"`
	ReleaseName     string `json:"release_name"`
	ChartName       string `json:"chart_name"`
	ValuesYaml      string `json:"values_yaml"`
	CreateNamespace bool   `json:"create_namespace"`
}

type Chart struct {
	Name         string       `yaml:"name"`
	Version      string       `yaml:"version"`
	Description  string       `yaml:"description"`
	URLs         []string     `yaml:"urls"`
	Digest       string       `yaml:"digest"`
	Dependencies []Dependency `yaml:"dependencies"`
}

// Dependency represents a dependency for a Helm chart
type Dependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// RepoIndex represents the index.yaml file structure
type RepoIndex struct {
	Entries map[string][]Chart `yaml:"entries"`
}

// ChartInfo represents the structure of the JSON response
type ChartInfo struct {
	Repository string  `json:"repository"`
	Charts     []Chart `json:"charts"`
}
