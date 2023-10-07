package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"WireguardManager/internal/config"
	"WireguardManager/internal/logging"
	storage "WireguardManager/internal/repository/sqlite"
	"WireguardManager/internal/service"
	"WireguardManager/internal/transport/rest"
	"WireguardManager/internal/transport/rest/handler"
	"WireguardManager/internal/utility/network"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	// Test CI changes jjj

	//
	// Preconfigure logger
	//
	logging.PreConfigure()

	//
	// Load new configuration from env variables and config files
	//
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatal("Config error: ", err.Error())
	}

	//
	// Configure logger with new configuration
	//
	err = logging.Configure(logging.Deps{
		ConsoleLogLevel: conf.Logging.ConsoleLevel,
		FileLogLevel:    conf.Logging.FileLevel,
		FilePath:        conf.Logging.FilePath,
	})
	if err != nil {
		logrus.Fatal("Logging configure error: ", err.Error())
	}

	//
	// Connect database
	//
	db, err := gorm.Open(storage.NewSqliteConnection(conf.Database.FilePath), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Db error:", err.Error())
	}

	//
	// Init NetworkTool for driving Wireguard interface and TrafficControl tool
	//
	netTool := network.NewNetworkTool(conf.Wireguard.Port)

	//
	// Create and prepare repositories
	//
	repositories := storage.NewRepositories(db)
	repositoriesInitDeps := storage.InitDeps{
		NetTool:          netTool,
		WireguardPort:    conf.Wireguard.Port,
		WireguardEnabled: true,
		PeerCount:        conf.Wireguard.PeerLimit,
	}
	if err = repositories.Init(repositoriesInitDeps); err != nil {
		logrus.Fatal("Repository error:", err.Error())
	}

	//
	// Init services
	//
	services := service.NewServices(service.Deps{
		NetTool:          netTool,
		PeerRepository:   repositories.PeerRepository,
		ServerRepository: repositories.ServerRepository,
	})
	if err := services.RecoverService.RecoverServer(); err != nil {
		logrus.Fatal("Failed to recover server: ", err.Error())
	}
	if err := services.RecoverService.RecoverPeers(); err != nil {
		logrus.Fatal("Failed to recover peers: ", err.Error())
	}

	//
	// Init REST API endpoint handlers
	//
	handler := handler.NewHandler(handler.Deps{
		PeerService:   services.PeerService,
		ServerService: services.ServerService,
		Configuration: *conf,
	}).Init(conf.RestApi.GinMode)

	//
	// Init REST API server
	//
	restServer := rest.NewServer(conf.RestApi.Port, handler)

	//
	// Run REST API server
	//
	go func() {
		logrus.Info("Starting REST api server")
		if err := restServer.ListenAndServe(); err != nil {
			logrus.Fatal("rest listen and serve error: ", err.Error())
		}
	}()

	//
	// Waiting for exit
	//
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logrus.Info("Shutting down server")

	//
	// Shutdown REST API server
	//
	if err = restServer.Stop(context.Background()); err != nil {
		logrus.Errorf("error occurred on rest server shutting down: %s", err.Error())
	}
}
