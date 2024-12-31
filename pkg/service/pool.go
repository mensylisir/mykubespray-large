package service

import (
	"errors"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type PoolService interface {
	CopyFile(srcFile, destFile string, hosts []entity.Host) error
	AddHosts(record entity.Record, hosts []entity.Host) error
	ExecuteCommand(command string, hosts []entity.Host) error
	AddDNS(dns string, hosts []entity.Host) error
}

type poolService struct {
}

func NewPoolService() poolService {
	return poolService{}
}

func (pool poolService) CopyFile(srcFile, destFile string, hosts []entity.Host) error {
	execPool := utils.NewSSHExecutorPool()
	result := execPool.CopyFileParallel(srcFile, destFile, hosts)
	if result.OverallSuccess {
		return nil
	}
	return errors.New("add keycloak cert failed")
}

func (pool poolService) AddHosts(record entity.Record, hosts []entity.Host) error {
	execPool := utils.NewSSHExecutorPool()
	result := execPool.AddHostsParallel(record, hosts)
	if result.OverallSuccess {
		return nil
	}
	return errors.New("add /etc/hosts failed")
}

func (pool poolService) AddDNS(dns string, hosts []entity.Host) error {
	execPool := utils.NewSSHExecutorPool()
	result := execPool.AddDNSParallel(dns, hosts)
	if result.OverallSuccess {
		return nil
	}
	return errors.New("add /etc/resolv.conf failed")
}

func (pool poolService) ExecuteCommand(command string, hosts []entity.Host) error {
	execPool := utils.NewSSHExecutorPool()
	result := execPool.ExecuteCommandParallel(command, hosts)
	if result.OverallSuccess {
		return nil
	}
	return errors.New("execute command failed")
}
