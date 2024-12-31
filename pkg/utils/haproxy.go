package utils

import (
	"bytes"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"strings"
	"text/template"
)

type HaproxyClient struct {
	HaproxyConf entity.HaproxyConf
	OSClient    OSClient
}

func NewHaproxyClient(haproxyConf entity.HaproxyConf, osClient OSClient) *HaproxyClient {
	return &HaproxyClient{
		HaproxyConf: haproxyConf,
		OSClient:    osClient,
	}
}

func (client *HaproxyClient) InstallHaproxy(logChan chan LogEntry) error {
	command := ""
	os, err := GetDistribution(&client.OSClient.SSExecutor)
	if err != nil {
		logger.GetLogger().Printf("Failed to create ssh connection: %s", err.Error())
		return err
	}
	if os == "ubuntu" {
		command = "sudo apt install haproxy -y"
	} else if os == "centos" {
		command = "sudo yum install haproxy -y"
	}
	err = client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Printf("Failed to install haproxy: %s", err.Error())
		return err
	}
	return nil
}

func (client *HaproxyClient) ConfigureHaproxy() error {
	configFile := "/etc/haproxy/haproxy.cfg"
	templateText := `
global
  log /dev/log  local0 warning
  chroot      /var/lib/haproxy
  pidfile     /var/run/haproxy.pid
  maxconn     4000
  user        haproxy
  group       haproxy
  daemon
  stats socket /var/lib/haproxy/stats
defaults
  log global
  option  httplog
  option  dontlognull
  timeout connect 5000
  timeout client 50000
  timeout server 50000
frontend kube-apiserver
  bind *:6443
  mode tcp
  option tcplog
  default_backend kube-apiserver
backend kube-apiserver
  mode tcp
  option tcplog
  option tcp-check
  balance roundrobin
  default-server inter 10s downinter 5s rise 2 fall 2 slowstart 60s maxconn 250 maxqueue 256 weight 100
  {{ .StrServers }}
	`
	for index, server := range client.HaproxyConf.Servers {
		client.HaproxyConf.StrServers += fmt.Sprintf("server kube-apiserver-%d %s check\n  ", index, server)
	}
	client.HaproxyConf.StrServers = strings.TrimSpace(client.HaproxyConf.StrServers)
	tmpl, err := template.New("haproxy.conf").Parse(templateText)
	if err != nil {
		logger.GetLogger().Printf("Failed to generate template object: %s", err.Error())
		return err
	}
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, client.HaproxyConf)
	if err != nil {
		logger.GetLogger().Printf("Failed to generate template: %s", err.Error())
		return err
	}
	command := fmt.Sprintf("bash -c \"echo '%s' > %s\"", rendered.String(), configFile)
	if client.OSClient.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.OSClient.SSExecutor.Host.Password)
	}
	err = client.OSClient.SSExecutor.ExecuteCommandWithoutReturn(command)
	if err != nil {
		logger.GetLogger().Printf("Failed to generate haproxy config: %s", err.Error())
		return err
	}
	return nil
}
