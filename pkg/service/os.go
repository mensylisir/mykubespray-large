package service

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type OSService interface {
	Mount(conf entity.DiskConf) error
	AddHost(conf entity.RecordConf) error
	CopyFile(conf entity.CertConf) error
}

type osService struct {
}

func NewOSService() osService {
	return osService{}
}

func (os osService) Mount(conf entity.DiskConf) error {
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
	client := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	data, err := client.QueryVGName()
	if err != nil {
		logger.GetLogger().Errorf("Failed to query pv: %s", err)
		return err
	}
	conf.LVS = *data
	err = client.CreatePV(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create pv: %s", err)
		return err
	}
	err = client.ExtendVG(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to extend vg: %s", err)
		return err
	}
	err = client.ExtendLVPercent100(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to extend lv: %s", err)
		return err
	}
	err = client.XGrowFS(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to xfs-growfs /dev/mapper/%s-%s: %s", conf.VGName, conf.LVName, err)
		return err
	}
	return nil
}

func (os osService) AddHost(conf entity.RecordConf) error {
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
	client := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	return client.AddHost(conf.Record)
}

func (os osService) CopyFile(conf entity.CertConf) error {
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
	client := utils.NewOSClient(osCOnf, *sshExecutor, *localExecutor)
	return client.CopyFile(conf.CertPath, conf.DestPath)
}
