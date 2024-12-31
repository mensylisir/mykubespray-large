package controller

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
)

type KubernetesController struct {
	Ctx               context.Context
	kubernetesService service.KubernetesService
}

func NewKubernetesController() *KubernetesController {
	return &KubernetesController{
		kubernetesService: service.NewKubernetesService(),
	}
}

var kubernetesController KubernetesController

func init() {
	kubernetesController = *NewKubernetesController()
}

func ApplyYAMLs(ctx *gin.Context) {
	var kubernetesConf entity.KubernetesFilesConf
	if err := ctx.ShouldBind(&kubernetesConf); err != nil {
		logger.GetLogger().Errorf("KubernetesFilesConf bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	results, err := kubernetesController.kubernetesService.ApplyYAMLs(kubernetesConf)
	if !results.OverallSuccess || err != nil {
		err := errors.New("Apply yaml failed")
		ginx.NewRender(ctx).Data(results, err)
	} else {
		ginx.NewRender(ctx).Data(results, nil)
	}
}

func AddRepo(ctx *gin.Context) {
	var helmRepository entity.HelmRepository
	if err := ctx.ShouldBind(&helmRepository); err != nil {
		logger.GetLogger().Errorf("HelmRepository bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := kubernetesController.kubernetesService.AddRepo(helmRepository)
	if err != nil {
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("success", nil)
}

func InstallChart(ctx *gin.Context) {
	var helmChartInfo entity.HelmChartInfo
	if err := ctx.ShouldBind(&helmChartInfo); err != nil {
		logger.GetLogger().Errorf("HelmChartInfo bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	data, err := kubernetesController.kubernetesService.InstallChart(helmChartInfo)
	if err != nil {
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data(data, nil)
}
