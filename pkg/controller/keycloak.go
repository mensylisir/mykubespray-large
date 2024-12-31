package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
)

type KeycloakController struct {
	Ctx             context.Context
	keycloakService service.KeycloakService
}

func NewKeycloakController() *KeycloakController {
	return &KeycloakController{
		keycloakService: service.NewKeycloakService(),
	}
}

var keycloakController KeycloakController

func init() {
	keycloakController = *NewKeycloakController()
}

func CreateGroup(ctx *gin.Context) {
	var groupConf entity.GroupConf
	if err := ctx.ShouldBind(&groupConf); err != nil {
		logger.GetLogger().Errorf("GroupConf bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	err := keycloakController.keycloakService.CreateGroup(groupConf)
	if err != nil {
		logger.GetLogger().Errorf("Create group failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data("Create group success", nil)
}

func QueryUserByName(ctx *gin.Context) {
	var userConf entity.UserConf
	if err := ctx.ShouldBind(&userConf); err != nil {
		logger.GetLogger().Errorf("GroupConf bind failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	data, err := keycloakController.keycloakService.QueryUserByName(userConf)
	if err != nil {
		logger.GetLogger().Errorf("Query User failed: %s", err.Error())
		ginx.Dangerous(err)
	}
	ginx.NewRender(ctx).Data(data, nil)
}
