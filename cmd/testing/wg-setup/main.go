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
	"WireguardManager/internal/tools/ipt"
	"WireguardManager/internal/tools/wg"

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
		logrus.Fatal("Failed to load config: ", err)
	}

	// Setup Wg interface

	wgTool, err := wg.NewTool()
	if err != nil {
		logrus.Fatal("Failed to create Tool: ", err)
	}

	if err := wgTool.AddWgInterface(wg.ServerParams{
		InfName:    WgInfName,
		Ip:         net.ParseIP(WgServerIp),
		Mask:       net.CIDRMask(WgServerMask, 32),
		PrivateKey: conf.ServerPrivateKey,
		Port:       conf.ServerPort,
	}); err != nil {
		logrus.Fatal("Failed to add wg interface: ", err)
	}

	logrus.Infof("Server enabled")

	if err := wgTool.AddWgPeer(wg.PeerParams{
		InfName:   WgInfName,
		Ip:        net.ParseIP(WgPeerIp),
		Mask:      net.CIDRMask(WgPeerMask, 32),
		PublicKey: conf.PeerPublicKey,
	}); err != nil {
		logrus.Fatal("Failed to add wg peer: ", err.Error())
	}

	logrus.Infof("Peer enabled")

	// Add NAT rules and check them
	iptablesTool, err := ipt.NewTool()
	if err != nil {
		logrus.Fatal("Failed to create iptables tool: ", err)
	}

	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))
	wgNetParams := ipt.NatParams{
		WgInf:  WgInfName,
		GwInf:  GwInfName,
		WgPort: conf.ServerPort,
		WgNet:  wgNetPrefix,
	}

	if err := iptablesTool.AddWgNatRules(wgNetParams); err != nil {
		logrus.Fatal("Failed to add wg NAT rules: ", err)
	}
	logrus.Infof("NAT enabled")

	if err := iptablesTool.CheckWgNatRules(wgNetParams); err != nil {
		logrus.Fatal("Failed to check wg NAT rules: ", err)
	}
	logrus.Infof("Wg nat exists")

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
