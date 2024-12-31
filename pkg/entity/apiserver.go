package entity

import v1 "k8s.io/api/core/v1"

type ApiServer struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Spec       v1.PodSpec `yaml:"spec"`
}

type ApiServerOidcConf struct {
	Host               Host
	OIDCIssuerUrl      string
	OIDCClientID       string
	OIDCUsernameClaim  string
	OIDCUsernamePrefix string
	OIDCGroupsClaim    string
	OIDCCAFile         string
}
