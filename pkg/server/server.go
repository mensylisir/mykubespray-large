package server

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/whoisfisher/mykubespray/pkg/httpx"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/router"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type Server struct {
	ConfigFile string
	Version    string
}

type ServerOption func(*Server)

func SetConfigFile(file string) ServerOption {
	return func(server *Server) {
		server.ConfigFile = file
	}
}

func SetVersion(version string) ServerOption {
	return func(server *Server) {
		server.Version = version
	}
}

func Run(options ...ServerOption) {
	code := 1
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	server := Server{
		ConfigFile: filepath.Join("pkg", "conf", "config.yaml"),
		Version:    "not specified",
	}

	for _, option := range options {
		option(&server)
	}
	server.ReadConfig()
	logger.Init()
	cleanFunc, err := server.initialize()
	if err != nil {
		fmt.Println("server init fail")
		os.Exit(code)
	}
EXIT:
	for {
		sgn := <-signalChannel
		fmt.Println("received signal:", sgn.String())
		switch sgn {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			code = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}
	cleanFunc()
	fmt.Println("server exit")
	os.Exit(code)
}

func (server Server) ReadConfig() error {
	if server.ConfigFile == "" {
		server.ConfigFile = "../conf/config.yaml"
	}
	viper.SetConfigFile(server.ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	return nil

}

func (server Server) initialize() (func(), error) {
	fns := Functions{}
	_, cancel := context.WithCancel(context.Background())
	fns.Add(cancel)
	route := router.New(server.Version)
	go func() {
		err := http.ListenAndServe(":6060", nil)
		if err != nil {
			logger.GetLogger().Errorf("Failed to bind 6060 debug info: %s", err.Error())
			return
		}
	}()
	httpClean := httpx.Init(route)
	fns.Add(httpClean)
	return fns.Ret(), nil
}

type Functions struct {
	List []func()
}

func (fs *Functions) Add(f func()) {
	fs.List = append(fs.List, f)
}

func (fs *Functions) Ret() func() {
	return func() {
		for i := 0; i < len(fs.List); i++ {
			fs.List[i]()
		}
	}
}
