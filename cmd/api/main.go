package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"WireguardManager/internal/config"
	"WireguardManager/internal/logging"
	"WireguardManager/internal/repositories/sqlite"
	"WireguardManager/internal/services"
	"WireguardManager/internal/tools/auth"
	"WireguardManager/internal/tools/network"
	"WireguardManager/internal/transport/rest"
	"WireguardManager/internal/transport/rest/handlers"

	"github.com/sirupsen/logrus"
)

func main() {

	//
	// 		PREPARE LOGGING AND CONFIGURATION
	//

	logging.SetTempConfiguration()

	conf, err := config.LoadConfiguration()
	if err != nil {
		logrus.Fatal("Config error: ", err.Error())
	}

	err = logging.Configure(logging.Deps{
		ConsoleLogLevel: conf.Logging.ConsoleLevel,
		FileLogLevel:    conf.Logging.FileLevel,
		FilePath:        conf.Logging.FilePath,
	})
	if err != nil {
		logrus.Fatal("Logging configure error: ", err.Error())
	}

	//
	// 		CREATE PARTS and INJECT DEPENDENCIES
	//

	// Connect to database
	db, err := sqlite.NewSqliteDb(conf.Database.FilePath)
	if err != nil {
		logrus.Fatal("Failed to connect db:", err.Error())
	}

	// Create repositories
	repositories := sqlite.NewRepositories(db)

	// Create NetworkTool for driving Wireguard interface and TrafficControl tool
	netTool := network.NewNetworkTool(conf.Wireguard.Port)

	// Create services
	services := services.NewServices(services.Deps{
		NetTool:          netTool,
		PeerRepository:   repositories.PeerRepository,
		ServerRepository: repositories.ServerRepository,
	})

	// Create jwt auth tool
	authTool := auth.NewJwtAuthTool()

	// Create REST api handlers
	handlers := handlers.NewHandler(handlers.Deps{
		PeerService:   services.PeerService,
		ServerService: services.ServerService,
		Configuration: *conf,
		AuthTool:      authTool,
	})

	// Create REST api server
	restServer := rest.NewServer(conf.RestApi.Port, handlers)

	//
	// 		PREPARE TO LAUNCH (INITIALIZATION)
	//

	// Init repositories
	if err = repositories.Init(sqlite.InitDeps{
		NetTool:          netTool,
		WireguardPort:    conf.Wireguard.Port,
		WireguardEnabled: true,
		PeerCount:        conf.Wireguard.PeerLimit,
	}); err != nil {
		logrus.Fatal("Repository error:", err.Error())
	}

	// Recover app runtime state from data models
	if err := services.RecoverService.RecoverAll(); err != nil {
		logrus.Fatal("Failed to recover state:", err.Error())
	}

	// Load jwt keys to auth tokens
	if err = authTool.LoadJwtKeys(auth.KeysDeps{
		HS256SecretKey:     conf.Jwt.HS256SecretKey,
		RS256PublicKeyPath: conf.Jwt.RS256PublicKeyPath,
	}); err != nil {
		logrus.Fatal("Failed to load jwt keys:", err.Error())
	}

	// Load SSL cert to REST https
	if err = restServer.LoadSSL(rest.SslDeps{
		CrtPath: conf.RestApi.Ssl.CrtPath,
		KeyPath: conf.RestApi.Ssl.KeyPath,
	}); err != nil {
		logrus.Fatal("Failed to load SSL:", err.Error())
	}

	//
	// 		LAUNCH APP
	//

	// Run REST api http server
	go func() {
		logrus.Info("Starting REST api server")
		if err := restServer.ListenAndServe(); err != nil {
			logrus.Fatal("Rest listen and serve error: ", err.Error())
		}
	}()

	//
	// 		WAITING FOR EXIT
	//

	// Wait signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logrus.Info("Shutting down server")

	// Shutdown REST api server
	if err = restServer.Stop(context.Background()); err != nil {
		logrus.Errorf("error occurred on rest server shutting down: %s", err.Error())
	}
}
