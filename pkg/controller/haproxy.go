package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
)

type HaproxyController struct {
	Ctx            context.Context
	haproxyService service.HaproxyService
}

func NewHaproxyController() *HaproxyController {
	return &HaproxyController{
		haproxyService: service.NewHaproxyService(),
	}
}

var haproxyController HaproxyController

func init() {
	haproxyController = *NewHaproxyController()
}

func ConfigureHaproxy(ctx *gin.Context) {
	var haproxyDTO entity.HaproxyConf
	if err := ctx.ShouldBind(&haproxyDTO); err != nil {
		logger.GetLogger().Errorf("haproxyconf bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := haproxyController.haproxyService.Configure(haproxyDTO)
	if err != nil {
		logger.GetLogger().Errorf("Configure haproxy failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Configure haproxy success", nil)
}
