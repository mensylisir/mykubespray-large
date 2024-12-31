package service

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type HaproxyService interface {
	Configure(conf entity.HaproxyConf) error
}

type haproxyService struct {
}

func NewHaproxyService() haproxyService {
	return haproxyService{}
}

func (hs haproxyService) Configure(conf entity.HaproxyConf) error {
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
	osCOnf := utils.OSConf{}
	localExecutor := utils.NewLocalExecutor()
	//sshExecutor := utils.NewSSHExecutor(*connection)
	sshExecutor := utils.NewExecutor(conf.Host)
	osclient := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	client := utils.NewHaproxyClient(conf, *osclient)
	err := client.ConfigureHaproxy()
	if err != nil {
		logger.GetLogger().Errorf("Failed to configure Haproxy: %s", err)
		return err
	}
	err = client.OSClient.DaemonReload()
	if err != nil {
		logger.GetLogger().Errorf("Failed to daemon reload: %s", err)
		return err
	}
	err = client.OSClient.StartService("haproxy")
	if err != nil {
		logger.GetLogger().Errorf("Failed to start haproxy: %s", err)
		return err
	}
	return nil
}
