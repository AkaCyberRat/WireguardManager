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
	"WireguardManager/internal/tools/ipt"
	"WireguardManager/internal/tools/tc"
	"WireguardManager/internal/tools/wg"

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
	Speed uint
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
	IfbInfName   = "ifb0"
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

	//
	// 		Server actions
	//

	// Setup Wg interface

	wgService, err := wg.NewWgTool()
	if err != nil {
		logrus.Fatal("Failed to create Tool: ", err.Error())
	}

	wgServerIp := net.ParseIP(WgServerIp)
	wgServerMask := net.CIDRMask(WgServerMask, 32)
	serverPrivateKey := conf.ServerPrivateKey
	port := conf.ServerPort

	if err := wgService.AddWgInterface(WgInfName, wgServerIp, wgServerMask, serverPrivateKey, port); err != nil {
		logrus.Fatal("Failed to setup Wg interface: ", err.Error())
	}
	logrus.Infof("Server enabled")

	// Add TC rules and check them
	tcTool, err := tc.NewTcTool()
	if err != nil {
		logrus.Fatal("Failed to create tc tool: ", err)
	}

	if err := tcTool.SetupTcBase(WgInfName, IfbInfName); err != nil {
		logrus.Fatal("Failed to setup tc base: ", err.Error())
	}
	logrus.Infof("Tc base enabled")

	if err := tcTool.CheckTcBaseRules(WgInfName, IfbInfName); err != nil {
		logrus.Fatal("Failed to check tc base rules existence: ", err)
	} else {
		logrus.Infof("Tc base rules exist")
	}

	// Add NAT rules and check them
	iptablesTool, err := ipt.NewIptablesTool()
	if err != nil {
		logrus.Fatal("Failed to create iptables tool: ", err)
	}

	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))
	wgNetParams := ipt.WgNatParams{
		WgInf:  WgInfName,
		GwInf:  GwInfName,
		WgPort: port,
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

	//
	// 		Peer actions
	//

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

		if err := tcTool.ApplyTcForPeer(tc.TcPeerParams{
			WgInf:             WgInfName,
			IfbInf:            IfbInfName,
			PeerIp:            wgPeerIp,
			ServerNetworkMask: wgServerMask,
			DownloadSpeedMb:   peer.Speed,
			UploadSpeedMb:     peer.Speed,
		}); err != nil {
			logrus.Fatal("Failed to apply tc rules for peer: ", err.Error())
		}
		logrus.Infof("Tc rules for peer %d enabled", i)

	}

	go StartTimedValidator(context.Background(), 30*time.Second, wgService, iptablesTool, tcTool, conf)

	// Wait for exit  signal
	waitForExitSignal()

	logrus.Info("Shutting down")

}

func StartTimedValidator(ctx context.Context, interval time.Duration, wgService wg.Tool, iptablesTool ipt.IptablesTool, tcTool tc.TcTool, config Configuration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			startTime := time.Now()
			err := ValidateOnce(wgService, iptablesTool, tcTool, config)
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

func ValidateOnce(wgService wg.Tool, iptablesTool ipt.IptablesTool, tcTool tc.TcTool, config Configuration) error {
	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))

	if err := wgService.CheckWgInterface(WgInfName); err != nil {
		return fmt.Errorf("wg interface check failed: %w", err)
	}

	if err := iptablesTool.CheckWgNatRules(ipt.WgNatParams{
		WgInf:  WgInfName,
		GwInf:  GwInfName,
		WgPort: config.ServerPort,
		WgNet:  wgNetPrefix,
	}); err != nil {
		return fmt.Errorf("wg nat check failed: %w", err)
	}

	if err := tcTool.CheckTcBaseRules(WgInfName, IfbInfName); err != nil {
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
