package service

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils/kubernetes"
	"helm.sh/helm/v3/pkg/release"
)

type KubernetesService interface {
	ApplyYAMLs(conf entity.KubernetesFilesConf) (*entity.ApplyResults, error)
	AddRepo(conf entity.HelmRepository) error
	InstallChart(conf entity.HelmChartInfo) (*release.Release, error)
}

type kubernetesService struct {
}

func NewKubernetesService() kubernetesService {
	return kubernetesService{}
}

func (ks kubernetesService) ApplyYAMLs(conf entity.KubernetesFilesConf) (*entity.ApplyResults, error) {
	client, err := kubernetes.NewK8sClient(conf.K8sConfig)
	if err != nil {
		logger.GetLogger().Errorf("Error creating kubernetes client: %v", err)
		return nil, err
	}
	results, err := client.ApplyYAMLs(conf.Files)
	if err != nil {
		logger.GetLogger().Errorf("Error apply files to kubernetes: %v", err)
		return nil, err
	}
	return results, nil
}

func (ks kubernetesService) AddRepo(conf entity.HelmRepository) error {
	client, err := kubernetes.NewK8sClient(conf.K8sConfig)
	if err != nil {
		logger.GetLogger().Errorf("Error creating kubernetes client: %v", err)
		return err
	}
	err = client.AddOrUpdateChartRepo(conf)
	if err != nil {
		logger.GetLogger().Errorf("Error Add helm repo to kubernetes: %v", err)
		return err
	}
	return nil
}

func (ks kubernetesService) InstallChart(conf entity.HelmChartInfo) (*release.Release, error) {
	client, err := kubernetes.NewK8sClient(conf.K8sConfig)
	if err != nil {
		logger.GetLogger().Errorf("Error creating kubernetes client: %v", err)
		return nil, err
	}
	data, err := client.InstallOrUpgradeChart(conf)
	if err != nil {
		logger.GetLogger().Errorf("Error Install helm charts to kubernetes: %v", err)
		return nil, err
	}
	return data, nil
}
