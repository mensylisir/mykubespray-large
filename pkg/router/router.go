package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/whoisfisher/mykubespray/pkg/aop"
	"github.com/whoisfisher/mykubespray/pkg/controller"
	"os"
)

func New(version string) *gin.Engine {

	PrintAccessLog := viper.GetBool("bind.print_access_log")
	RunMode := viper.GetString("app.run_mode")
	gin.SetMode(RunMode)

	loggerMid := aop.Logrus()
	recoveryMid := aop.Recovery()
	r := gin.New()
	r.Use(recoveryMid)
	if PrintAccessLog {
		r.Use(loggerMid)
	}
	r.Use(aop.LogRequest())
	r.Use(aop.LoggerMiddleware())
	configRoute(r, version)
	return r
}

func configRoute(r *gin.Engine, version string) {
	httpRouter := r.Group("/api/v1")
	configHttpRouter(httpRouter, version)
	//
	ws := r.Group("/api/ws/v1")
	configWebsocketRouter(ws)
}

func configWebsocketRouter(rg *gin.RouterGroup) {
	rg.Use(aop.Cors())
	rg.GET("/cluster/create", controller.CreateCluster)
	rg.GET("/cluster/delete", controller.DeleteCluster)
	rg.GET("/cluster/nodes/add", controller.AddNodeToCluster)
	rg.GET("/cluster/node/delete", controller.DeleteNodeFromCluster)
}

func configHttpRouter(rg *gin.RouterGroup, version string) {
	rg.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	rg.GET("/pid", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", os.Getpid()))
	})
	rg.GET("/addr", func(c *gin.Context) {
		c.String(200, c.Request.RemoteAddr)
	})
	rg.GET("/version", func(c *gin.Context) {
		c.String(200, version)
	})
	rg.POST("/keepalived/configure", controller.ConfigureKeepalived)
	rg.POST("/haproxy/configure", controller.ConfigureHaproxy)
	rg.POST("/apiserver/configure", controller.ConfigureApiServer)
	rg.POST("/server/mount", controller.MountDisk)
	rg.POST("/server/hosts", controller.AddHosts)
	rg.POST("/server/cert/copy", controller.CopyFile)
	rg.POST("/server/cert/copyparallel", controller.CopyFileParallel)
	rg.POST("/server/hostsparallel", controller.AddHostsParallel)
	rg.POST("/server/dnsparallel", controller.AddDNSParallel)
	rg.POST("/server/execmdparallel", controller.ExecuteCommandParallel)
	rg.POST("/keycloak/group", controller.CreateGroup)
	rg.POST("/keycloak/user", controller.QueryUserByName)
	rg.POST("/kubernetes/apply", controller.ApplyYAMLs)
	rg.POST("/kubernetes/helm/repo", controller.AddRepo)
	rg.POST("/kubernetes/helm/chart", controller.InstallChart)
}
