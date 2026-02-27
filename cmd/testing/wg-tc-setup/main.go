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
	Peers            []Peer
}

type Peer struct {
	Ip    string
	Speed int
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
	const WgPeerMask = 32

	logging.SetTempConfiguration()

	logrus.Info("Starting wg-tc-setup")

	// Load configuration

	configPath := getConfigPathFromArgs()
	conf, err := config.LoadConfiguration[Configuration](configPath)
	if err != nil {
		logrus.Fatal("Failed to load config: ", err.Error())
	}

	// Setup Wg interface

	wgService, err := network.NewWgService()
	if err != nil {
		logrus.Fatal("Failed to create WgService: ", err.Error())
	}

	wgServerIp := net.ParseIP(WgServerIp)
	wgServerMask := net.CIDRMask(WgServerMask, 32)
	serverPrivateKey := conf.ServerPrivateKey
	port := conf.ServerPort

	if err := wgService.SetupWgInterface(WgInfName, wgServerIp, wgServerMask, serverPrivateKey, port); err != nil {
		logrus.Fatal("Failed to setup Wg interface: ", err.Error())
	}
	logrus.Infof("Server enabled")

	// Setup TC base rules
	if err := network.SetupTcBase(WgInfName); err != nil {
		logrus.Fatal("Failed to setup tc base: ", err.Error())
	}
	logrus.Infof("Tc base enabled")

	// Check TC base rules
	if err := network.CheckTcBase(WgInfName); err != nil {
		logrus.Fatal("Failed to check tc base rules existence: ", err)
	} else {
		logrus.Infof("Tc base rules exist")
	}

	// Setup NAT for wg interface

	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))

	if err := wgService.SetupWgNAT(WgInfName, GwInfName, port, wgNetPrefix); err != nil {
		logrus.Fatal("Failed to setup WgNAT: ", err.Error())
	}
	logrus.Infof("NAT enabled")

	// Check NAT for wg interface

	exists, err := wgService.IsWgNatExists(WgInfName, GwInfName, port, wgNetPrefix)
	if err != nil {
		logrus.Fatal("Failed to check NAT rules existanse: ", err)
	}
	logrus.Infof("Is NAT exists: %v", exists)

	wgPeerMask := net.CIDRMask(WgPeerMask, 32)
	peerPublicKey := conf.PeerPublicKey

	for i, peer := range conf.Peers {
		// Setup Wg peer

		wgPeerIp := net.ParseIP(peer.Ip)

		if err := wgService.AddWgPeer(WgInfName, wgPeerIp, wgPeerMask, peerPublicKey, nil); err != nil {
			logrus.Fatal("Failed to add Wg peer: ", err.Error())
		}
		logrus.Infof("Peer %d enabled", i)

		// Setup Tc for peer

		if err := network.ApplyTcForPeer(WgInfName, wgPeerIp, wgServerMask, peer.Speed, peer.Speed); err != nil {
			logrus.Fatal("Failed to apply tc rules for peer: ", err.Error())
		}
		logrus.Infof("Tc rules for peer %d enabled", i)

	}

	// Wait for exit  signal
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
