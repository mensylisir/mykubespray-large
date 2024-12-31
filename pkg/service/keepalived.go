package service

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type KeepalivedService interface {
	Configure(conf entity.KeepalivedConf) error
}

type keepalivedService struct {
}

func NewKeepalivedService() KeepalivedService {
	return keepalivedService{}
}

func (ks keepalivedService) Configure(conf entity.KeepalivedConf) error {
	sshConfig := utils.SSHConfig{}
	sshConfig.Host = conf.Host.Address
	sshConfig.Port = conf.Host.Port
	sshConfig.User = conf.Host.User
	sshConfig.Password = conf.Host.Password
	sshConfig.PrivateKey = conf.Host.PrivateKey
	sshConfig.AuthMethods = conf.Host.AuthMethods
	//connection, err := utils.NewSSHConnection(sshConfig)
	//if err != nil {
	//	logger.GetLogger().Errorf("Failed to create SSH connection: %s", err)
	//	return err
	//}
	conf.SrcIP = conf.Host.Address
	osCOnf := utils.OSConf{}
	localExecutor := utils.NewLocalExecutor()
	//sshExecutor := utils.NewSSHExecutor(*connection)
	sshExecutor := utils.NewExecutor(conf.Host)
	osclient := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	conf.IntFace = osclient.GetSpecifyNetCard(conf.Host.Address)
	client := utils.NewKeepalivedClient(conf, *osclient)
	err := client.ConfigureKeepalived()
	if err != nil {
		logger.GetLogger().Errorf("Failed to configure Keepalived: %s", err)
		return err
	}
	err = client.OSClient.DaemonReload()
	if err != nil {
		logger.GetLogger().Errorf("Failed to daemon reload: %s", err)
		return err
	}
	err = client.OSClient.StartService("keepalived")
	if err != nil {
		logger.GetLogger().Errorf("Failed to start keepalived: %s", err)
		return err
	}
	return nil
}
