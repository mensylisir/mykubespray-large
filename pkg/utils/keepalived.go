package utils

import (
	"bytes"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"strings"
	"text/template"
)

type KeepalivedClient struct {
	KeepalivedConf entity.KeepalivedConf
	OSClient       OSClient
}

func NewKeepalivedClient(KeepalivedConf entity.KeepalivedConf, osClient OSClient) *KeepalivedClient {
	return &KeepalivedClient{
		KeepalivedConf: KeepalivedConf,
		OSClient:       osClient,
	}
}

func (client *KeepalivedClient) InstallKeepalived(logChan chan LogEntry) error {
	command := ""
	os, err := GetDistribution(&client.OSClient.SSExecutor)
	if err != nil {
		logger.GetLogger().Printf("Failed to create ssh connection: %s", err.Error())
		return err
	}
	if os == "ubuntu" {
		command = "sudo apt install keepalived -y"
	} else if os == "centos" {
		command = "sudo yum install keepalived -y"
	}
	err = client.OSClient.SSExecutor.ExecuteCommand(command, logChan)
	if err != nil {
		logger.GetLogger().Printf("Failed to install keepalived: %s", err.Error())
		return err
	}
	return nil
}

func (client *KeepalivedClient) ConfigureKeepalived() error {
	configFile := "/etc/keepalived/keepalived.conf"
	templateText := `
global_defs {
  notification_email {
  }
  router_id LVS_DEVEL
  vrrp_skip_check_adv_addr
  vrrp_garp_interval 0
  vrrp_gna_interval 0
}
vrrp_script chk_haproxy {
  script "killall -0 haproxy"
  interval 2
  weight 2
}
vrrp_instance keepalived-vip {
  state {{ .State }}
  priority {{ .Priority }}
  interface {{ .IntFace }}
  virtual_router_id 51
  advert_int 1
  authentication {
    auth_type {{ .AuthType }}
    auth_pass {{ .AuthPass }}
  }
  unicast_src_ip {{ .SrcIP }}
  unicast_peer {
    {{ .StrPeers }}
  }
  virtual_ipaddress {
    {{ .VIP }}
  }
  track_script {
    chk_haproxy
  }
}
	`
	for _, peer := range client.KeepalivedConf.Peers {
		client.KeepalivedConf.StrPeers += fmt.Sprintf("%s\n    ", peer)
	}
	client.KeepalivedConf.StrPeers = strings.TrimSpace(client.KeepalivedConf.StrPeers)
	tmpl, err := template.New("Keepalived.conf").Parse(templateText)
	if err != nil {
		logger.GetLogger().Printf("Failed to generate template object: %s", err.Error())
		return err
	}
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, client.KeepalivedConf)
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
		logger.GetLogger().Printf("Failed to generate Keepalived config: %s", err.Error())
		return err
	}
	return nil
}

func (client *KeepalivedClient) IsVirtualIPActive() bool {
	command := "ip addr show dev" + client.KeepalivedConf.IntFace
	output, err := client.OSClient.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Printf("Failed to query Keepalived vip: %s", err.Error())
		return false
	}
	return strings.Contains(output, client.KeepalivedConf.VIP)
}
