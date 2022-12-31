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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	//
	// Set temp logging configuration
	//
	logging.TempConfig()

	//
	// Load new configuration from env variables and config files
	//
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatal("Config error: ", err.Error())
	}

	//
	// Reconfigure logging using new configuration
	//
	logging.Configure(logging.Deps{
		ConsoleLogLevel: conf.LoggingConsoleLevel,
		FileLogLevel:    conf.LoggingFileLevel,
		FilePath:        conf.LoggingFilePath,
	})

	//
	// Open database
	//
	db, err := gorm.Open(sqlite.Open(conf.DataBasePath), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Db err: %s", err.Error())
	}

	//
	// Create and initialize (if need) repositories
	//
	repositories, err := storage.NewRepositories(db).Init(storage.InitDeps{
		PeerCount: conf.WireguardPeerLimit,
	})
	if err != nil {
		logrus.Fatal("Repository err: ", err.Error())
	}

	//
	// Init NetworkTool for driving Wireguard interface and TrafficControl tool
	//
	netTool, err := network.NewNetworkTool(network.Deps{
		WireguardInterface: "wg0",
		WireguardIpNet:     "10.0.0.0/8",
		PrivateKey:         "uKIkAl5agqGLoodeDAdtgZHh91vXck5z/mmxETx2dWs=",
		Port:               conf.WireguardPort,
		UseTC:              true,
	})
	if err != nil {
		logrus.Fatal("Network error: ", err.Error())
	}

	//
	// Init services
	//
	services := service.NewServices(service.Deps{
		PeerRepos:          repositories.PeerRepos,
		TransactionManager: repositories.TransactionManager,
		NetTool:            netTool,
	})

	//
	// Init REST API endpoint handlers
	//
	handler := handler.NewHandler(handler.Deps{
		PeerService: services.PeerService,
	}).Init()

	//
	// Init REST API server
	//
	restServer := rest.NewServer(conf.ApiPort, handler)

	//
	// Run REST API server
	//
	go func() {
		logrus.Info("Starting REST api server")
		if err := restServer.ListenAndServe(conf.ApiPort); err != nil {
			logrus.Fatal("rest listen and serve err: ", err.Error())
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
