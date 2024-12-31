package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
)

type PoolController struct {
	Ctx         context.Context
	poolService service.PoolService
}

func NewPoolController() *PoolController {
	return &PoolController{
		poolService: service.NewPoolService(),
	}
}

var poolController PoolController

func init() {
	poolController = *NewPoolController()
}

func AddDNSParallel(ctx *gin.Context) {
	var addDNSParallel entity.AddDNSParallel
	if err := ctx.ShouldBind(&addDNSParallel); err != nil {
		logger.GetLogger().Errorf("AddDNSParallel bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := poolController.poolService.AddDNS(addDNSParallel.DNS, addDNSParallel.Hosts)
	if err != nil {
		logger.GetLogger().Errorf("Add /etc/resolv.conf failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Add /etc/resolv.conf success", nil)
}

func AddHostsParallel(ctx *gin.Context) {
	var addHostsParallel entity.AddHostsParallel
	if err := ctx.ShouldBind(&addHostsParallel); err != nil {
		logger.GetLogger().Errorf("AddHostsParallel bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := poolController.poolService.AddHosts(addHostsParallel.Record, addHostsParallel.Hosts)
	if err != nil {
		logger.GetLogger().Errorf("Add /etc/hosts failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Add /etc/hosts success", nil)
}

func CopyFileParallel(ctx *gin.Context) {
	var copyFileParallel entity.CopyFileParallel
	if err := ctx.ShouldBind(&copyFileParallel); err != nil {
		logger.GetLogger().Errorf("CopyFileParallel bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := poolController.poolService.CopyFile(copyFileParallel.SrcFile, copyFileParallel.DestFile, copyFileParallel.Hosts)
	if err != nil {
		logger.GetLogger().Errorf("Copy keycloak certificate failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Copy keycloak certificate success", nil)
}

func ExecuteCommandParallel(ctx *gin.Context) {
	var commandParallel entity.CommandParallel
	if err := ctx.ShouldBind(&commandParallel); err != nil {
		logger.GetLogger().Errorf("CommandParallel bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := poolController.poolService.ExecuteCommand(commandParallel.Command, commandParallel.Hosts)
	if err != nil {
		logger.GetLogger().Errorf("Execute command failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Execute command success", nil)
}
