package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

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

const (
	GwInfName    = "eth0"
	WgInfName    = "wg0"
	WgServerIp   = "11.0.0.1"
	WgServerMask = 24
	WgPeerMask   = 32
)

func main() {

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

	if err := wgService.CheckWgNat(WgInfName, GwInfName, port, wgNetPrefix); err != nil {
		logrus.Fatal("Failed to check NAT rules existence: ", err)
	}
	logrus.Infof("Wg nat exists")

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

	go StartTimedValidator(context.Background(), 30*time.Second, wgService, conf)

	// Wait for exit  signal
	waitForExitSignal()

	logrus.Info("Shutting down")

}

func StartTimedValidator(ctx context.Context, interval time.Duration, wgService network.WgService, config Configuration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			startTime := time.Now()
			err := ValidateOnce(wgService, config)
			duration := time.Since(startTime)

			if err != nil {
				logrus.Errorf("Timed validation failed (%v): %s", duration, err.Error())
			} else {
				logrus.Infof("Timed validation completed (%v)", duration)
			}
		case <-ctx.Done():
			log.Println("Timed validator stopped")
			return
		}
	}
}

func ValidateOnce(wgService network.WgService, config Configuration) error {
	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))

	if err := wgService.CheckWgInterface(WgInfName); err != nil {
		return fmt.Errorf("wg interface check failed: %w", err)
	}

	if err := wgService.CheckWgNat(WgInfName, GwInfName, config.ServerPort, wgNetPrefix); err != nil {
		return fmt.Errorf("wg nat check failed: %w", err)
	}

	if err := network.CheckTcBase(WgInfName); err != nil {
		return fmt.Errorf("tc validation failed: %w", err)
	}

	return nil
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
