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
	logging.PreConfigure()

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
		ConsoleLogLevel: conf.App.Logging.ConsoleLevel,
		FileLogLevel:    conf.App.Logging.FileLevel,
		FilePath:        conf.App.Logging.FolderPath,
	})

	//
	// Open database
	//
	db, err := gorm.Open(sqlite.Open(conf.App.Database.Path), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Db err: %s", err.Error())
	}

	//
	// Create and initialize (if need) repositories
	//
	repositories, err := storage.NewRepositories(db).Init(storage.InitDeps{
		PeerCount: conf.App.Wireguard.PeerLimit,
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
		Port:               conf.App.Wireguard.Port,
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

	if conf.App.LaunchMode == config.Develop {
		services.Init(service.InitDeps{
			PeerInitDeps: service.PeerInitDeps{PeersToCreate: conf.Develop.Services.Peer.PeersToCreate},
		})
	}

	//
	// Init REST API endpoint handlers
	//
	handler := handler.NewHandler(handler.Deps{
		PeerService: services.PeerService,
	}).Init(conf.App.RestApi.GinMode)

	//
	// Init REST API server
	//
	restServer := rest.NewServer(conf.App.RestApi.Port, handler)

	//
	// Run REST API server
	//
	go func() {
		logrus.Info("Starting REST api server")
		if err := restServer.ListenAndServe(); err != nil {
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
