package service

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type KubekeyService interface {
	//GenerateConfig() error
	CreateCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error
	DeleteCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error
	AddNodeToCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error
	DeleteNodeFromCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error
}

type kubekeyService struct {
}

func NewKubekeyService() kubekeyService {
	return kubekeyService{}
}

func (ks kubekeyService) CreateCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error {
	sshConfig := utils.SSHConfig{}
	registryHost := entity.Host{}
	for _, host := range conf.Hosts {
		if host.Registry != nil {
			sshConfig.Host = host.Address
			sshConfig.Port = host.Port
			sshConfig.User = host.User
			sshConfig.Password = host.Password
			sshConfig.PrivateKey = host.PrivateKey
			sshConfig.AuthMethods = host.AuthMethods
			conf.Registry = *host.Registry
			registryHost = host
		}
	}
	//connection, err := utils.NewSSHConnection(sshConfig)
	//if err != nil {
	//	logger.GetLogger().Errorf("Failed to create SSH connection: %s", err)
	//	return err
	//}
	osCOnf := utils.OSConf{}
	localExecutor := utils.NewLocalExecutor()
	//sshExecutor := utils.NewSSHExecutor(*connection)
	sshExecutor := utils.NewExecutor(registryHost)
	osclient := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	client := utils.NewKubekeyClient(conf, *osclient)
	if len(conf.VIPServer) > 0 {
		client.GenerateConfigWithVIP()
	} else {
		client.GenerateConfig()
	}
	client.CreateCluster(logChan)
	return nil
}

func (ks kubekeyService) DeleteCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error {
	sshConfig := utils.SSHConfig{}
	registryHost := entity.Host{}
	for _, host := range conf.Hosts {
		if host.Registry != nil {
			sshConfig.Host = host.Address
			sshConfig.Port = host.Port
			sshConfig.User = host.User
			sshConfig.Password = host.Password
			sshConfig.PrivateKey = host.PrivateKey
			sshConfig.AuthMethods = host.AuthMethods
			conf.Registry = *host.Registry
			registryHost = host
		}
	}
	//connection, err := utils.NewSSHConnection(sshConfig)
	//if err != nil {
	//	logger.GetLogger().Errorf("Failed to create SSH connection: %s", err)
	//	return err
	//}
	osCOnf := utils.OSConf{}
	localExecutor := utils.NewLocalExecutor()
	//sshExecutor := utils.NewSSHExecutor(*connection)
	sshExecutor := utils.NewExecutor(registryHost)
	osclient := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	client := utils.NewKubekeyClient(conf, *osclient)
	if len(conf.VIPServer) > 0 {
		client.GenerateConfigWithVIP()
	} else {
		client.GenerateConfig()
	}
	client.DeleteCluster(logChan)
	return nil
}

func (ks kubekeyService) AddNodeToCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error {
	sshConfig := utils.SSHConfig{}
	registryHost := entity.Host{}
	for _, host := range conf.Hosts {
		if host.Registry != nil {
			sshConfig.Host = host.Address
			sshConfig.Port = host.Port
			sshConfig.User = host.User
			sshConfig.Password = host.Password
			sshConfig.PrivateKey = host.PrivateKey
			sshConfig.AuthMethods = host.AuthMethods
			conf.Registry = *host.Registry
			registryHost = host
		}
	}
	//connection, err := utils.NewSSHConnection(sshConfig)
	//if err != nil {
	//	logger.GetLogger().Errorf("Failed to create SSH connection: %s", err)
	//	return err
	//}
	osCOnf := utils.OSConf{}
	localExecutor := utils.NewLocalExecutor()
	//sshExecutor := utils.NewSSHExecutor(*connection)
	sshExecutor := utils.NewExecutor(registryHost)
	osclient := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	client := utils.NewKubekeyClient(conf, *osclient)
	if len(conf.VIPServer) > 0 {
		client.GenerateConfigWithVIP()
	} else {
		client.GenerateConfig()
	}
	client.AddNode(logChan)
	return nil
}

func (ks kubekeyService) DeleteNodeFromCluster(conf entity.KubekeyConf, logChan chan utils.LogEntry) error {
	deleteNode := ""
	sshConfig := utils.SSHConfig{}
	registryHost := entity.Host{}
	for _, host := range conf.Hosts {
		if host.Registry != nil {
			sshConfig.Host = host.Address
			sshConfig.Port = host.Port
			sshConfig.User = host.User
			sshConfig.Password = host.Password
			sshConfig.PrivateKey = host.PrivateKey
			sshConfig.AuthMethods = host.AuthMethods
			conf.Registry = *host.Registry
			registryHost = host
		}
		if host.IsDeleted {
			deleteNode = host.Name
		}
	}
	//connection, err := utils.NewSSHConnection(sshConfig)
	//if err != nil {
	//	logger.GetLogger().Errorf("Failed to create SSH connection: %s", err)
	//	return err
	//}
	osCOnf := utils.OSConf{}
	localExecutor := utils.NewLocalExecutor()
	//sshExecutor := utils.NewSSHExecutor(*connection)
	sshExecutor := utils.NewExecutor(registryHost)
	osclient := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	client := utils.NewKubekeyClient(conf, *osclient)
	if len(conf.VIPServer) > 0 {
		client.GenerateConfigWithVIP()
	} else {
		client.GenerateConfig()
	}
	client.DeleteNode(deleteNode, logChan)
	return nil
}
