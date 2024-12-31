package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
)

type ApiServerController struct {
	Ctx              context.Context
	apiServerService service.ApiServeryService
}

func NewApiServerController() *ApiServerController {
	return &ApiServerController{
		apiServerService: service.NewApiServerService(),
	}
}

var apiServerController ApiServerController

func init() {
	apiServerController = *NewApiServerController()
}

func ConfigureApiServer(ctx *gin.Context) {
	var apiServerConf entity.ApiServerOidcConf
	if err := ctx.ShouldBind(&apiServerConf); err != nil {
		logger.GetLogger().Errorf("apiserverconf bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := apiServerController.apiServerService.ConfigureManifest(apiServerConf)
	if err != nil {
		logger.GetLogger().Errorf("Configure apiserver failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Configure apiserver success", nil)
}
