package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/whoisfisher/mykubespray/pkg/aop"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/service"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

type KubekeyController struct {
	Ctx            context.Context
	kubekeyService service.KubekeyService
}

func NewKubekeyController() *KubekeyController {
	return &KubekeyController{
		kubekeyService: service.NewKubekeyService(),
	}
}

var kubekeyController KubekeyController

func init() {
	kubekeyController = *NewKubekeyController()
}

func CreateCluster(ctx *gin.Context) {
	var conf entity.KubekeyConf
	ws, err := aop.UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.GetLogger().Errorf("Create websocket channel failed: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	err = ws.ReadJSON(&conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to read kkconf info: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	logChan := make(chan utils.LogEntry)
	go func() {
		for logEntry := range logChan {
			if logEntry.IsError {
				logger.GetLogger().Errorf("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			} else {
				logger.GetLogger().Infof("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			}
		}
	}()
	kubekeyController.kubekeyService.CreateCluster(conf, logChan)
}

func DeleteCluster(ctx *gin.Context) {
	var conf entity.KubekeyConf
	ws, err := aop.UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.GetLogger().Errorf("Create websocket channel failed: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	err = ws.ReadJSON(&conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to read kkconf info: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	logChan := make(chan utils.LogEntry)
	go func() {
		for logEntry := range logChan {
			if logEntry.IsError {
				logger.GetLogger().Errorf("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			} else {
				logger.GetLogger().Infof("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			}
		}
	}()
	kubekeyController.kubekeyService.DeleteCluster(conf, logChan)
}

func AddNodeToCluster(ctx *gin.Context) {
	var conf entity.KubekeyConf
	ws, err := aop.UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.GetLogger().Errorf("Create websocket channel failed: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	err = ws.ReadJSON(&conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to read kkconf info: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	logChan := make(chan utils.LogEntry)
	go func() {
		for logEntry := range logChan {
			if logEntry.IsError {
				logger.GetLogger().Errorf("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			} else {
				logger.GetLogger().Infof("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			}
		}
	}()
	kubekeyController.kubekeyService.AddNodeToCluster(conf, logChan)
}

func DeleteNodeFromCluster(ctx *gin.Context) {
	var conf entity.KubekeyConf
	ws, err := aop.UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.GetLogger().Errorf("Create websocket channel failed: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	err = ws.ReadJSON(&conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to read kkconf info: %s", err.Error())
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	logChan := make(chan utils.LogEntry)
	go func() {
		for logEntry := range logChan {
			if logEntry.IsError {
				logger.GetLogger().Errorf("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			} else {
				logger.GetLogger().Infof("%s\n", logEntry.Message)
				ws.WriteMessage(websocket.TextMessage, []byte(logEntry.Message))
			}
		}
	}()
	kubekeyController.kubekeyService.DeleteNodeFromCluster(conf, logChan)
}
