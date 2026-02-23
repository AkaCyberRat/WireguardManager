package main

import (
	"flag"
	"fmt"
	"net"
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
	const GwInfName, WgInfName = "eth0", "wg0"
	const WgServerIp, WgServerMask = "11.0.0.1", 24
	const WgPeerIp, WgPeerMask = "11.0.0.2", 32

	logging.SetTempConfiguration()

	logrus.Info("Starting wg-setup")

	// Load configuration

	configPath := getConfigPathFromArgs()
	conf, err := config.LoadConfiguration[Configuration](configPath)
	if err != nil {
		logrus.Fatal("Failed to load config: ", err.Error())
	}

	// Setup Wg interface

	wgServerIp := net.ParseIP(WgServerIp)
	wgServerMask := net.CIDRMask(WgServerMask, 32)
	serverPrivateKey := conf.ServerPrivateKey
	port := conf.ServerPort

	if err := network.SetupWgInterface(WgInfName, wgServerIp, wgServerMask, serverPrivateKey, port); err != nil {
		logrus.Fatal("Failed to setup Wg interface: ", err.Error())
	}
	logrus.Infof("Server enabled")

	// Setup Wg peer

	wgPeerIp := net.ParseIP(WgPeerIp)
	wgPeerMask := net.CIDRMask(WgPeerMask, 32)
	peerPublicKey := conf.PeerPublicKey

	if err := network.AddWgPeer(WgInfName, wgPeerIp, wgPeerMask, peerPublicKey, nil); err != nil {
		logrus.Fatal("Failed to add Wg peer: ", err.Error())
	}
	logrus.Infof("Peer enabled")

	// Setup NAT for wg interface

	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))

	if err := network.SetupWgNAT(WgInfName, GwInfName, port, wgNetPrefix); err != nil {
		logrus.Fatal("Failed to setup WgNAT: ", err.Error())
	}
	logrus.Infof("NAT enabled")

	// Check NAT for wg interface

	exists, err := network.IsWgNatExists(WgInfName, GwInfName, port, wgNetPrefix)
	if err != nil {
		logrus.Fatal("Failed to check NAT rules existanse: ", err)
	}
	logrus.Infof("Is NAT exists: ", exists)

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
