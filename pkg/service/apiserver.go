package service

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type ApiServeryService interface {
	ConfigureManifest(conf entity.ApiServerOidcConf) error
}

type apiServerService struct {
}

func NewApiServerService() apiServerService {
	return apiServerService{}
}

func (as apiServerService) ConfigureManifest(conf entity.ApiServerOidcConf) error {
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
	client := utils.NewApiServerClient(conf, *osclient)
	err := client.ModifyConfig()
	if err != nil {
		logger.GetLogger().Errorf("Failed to configure apiserver %s: %s", conf.Host.Name, err)
		return err
	}
	return nil
}
