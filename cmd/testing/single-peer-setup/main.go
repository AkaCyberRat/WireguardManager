package main

import (
	"WireguardManager/internal/core"
	"flag"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"WireguardManager/internal/config"
	"WireguardManager/internal/logging"
	"WireguardManager/internal/tools/network"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	PeerPublicKey    string
	ServerPrivateKey string
	ServerPort       int
}

func (c Configuration) ToDefault() config.Configuration {
	c.ServerPort = 51820
	return c
}

func (c Configuration) Validate() error {
	return nil
}

func main() {
	logging.SetTempConfiguration()

	configPath := getConfigPathFromArgs()

	conf, err := config.LoadConfiguration[Configuration](configPath)
	if err != nil {
		logrus.Fatal("Failed to load config: ", err.Error())
	}
	logrus.Infof("Configuration: %+v", conf)

	netTool := network.NewNetworkTool(conf.ServerPort)

	err = netTool.EnableServer(&core.Server{PublicKey: "", PrivateKey: conf.ServerPrivateKey, Enabled: true})
	if err != nil {
		logrus.Fatal("Failed to enable server: ", err.Error())
	}
	logrus.Infof("Server enabled")

	err = netTool.EnablePeer(&core.Peer{
		Id:            "2",
		Ip:            "11.0.0.2",
		PublicKey:     conf.PeerPublicKey,
		PresharedKey:  "",
		DownloadSpeed: 100,
		UploadSpeed:   100,
		Status:        core.Enabled,
	})
	if err != nil {
		logrus.Fatal("Failed to enable peer:", err.Error())
	}

	logrus.Infof("Peer enabled")

	wgNet := netip.MustParsePrefix("11.0.0.0/24")
	wgPort := uint16(conf.ServerPort)

	if err := network.SetupWgNAT("wg0", "eth0", wgPort, wgNet); err != nil {
		logrus.Fatal("Failed to setup WgNAT: ", err.Error())
	}

	logrus.Infof("WgNaT setup completed")

	// Wait for exit signal
	waitForExitSignal()

	logrus.Info("Shutting down")
}

func waitForExitSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}

func getConfigPathFromArgs() string {
	path := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()
	return *path
}
