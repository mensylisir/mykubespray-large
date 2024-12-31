package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
)

type KeepalivedController struct {
	Ctx               context.Context
	keepalivedService service.KeepalivedService
}

func NewKeepalivedController() *KeepalivedController {
	return &KeepalivedController{
		keepalivedService: service.NewKeepalivedService(),
	}
}

var keepalivedController KeepalivedController

func init() {
	keepalivedController = *NewKeepalivedController()
}

func ConfigureKeepalived(ctx *gin.Context) {
	var keepalivedDTO entity.KeepalivedConf
	if err := ctx.ShouldBind(&keepalivedDTO); err != nil {
		logger.GetLogger().Errorf("KeepalivedConf bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := keepalivedController.keepalivedService.Configure(keepalivedDTO)
	if err != nil {
		logger.GetLogger().Errorf("Configure keepalived failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Configure keepalived success", nil)
}
