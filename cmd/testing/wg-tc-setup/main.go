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
	// Make tools
	//
	wgTool, err := wg.NewTool()
	if err != nil {
		logrus.Fatal("Failed to create wg tool: ", err)
	}

	iptablesTool, err := ipt.NewTool()
	if err != nil {
		logrus.Fatal("Failed to create iptables tool: ", err)
	}

	tcTool, err := tc.NewTool()
	if err != nil {
		logrus.Fatal("Failed to create tc tool: ", err)
	}

	//
	// 		Server actions
	//

	if err := wgTool.AddWgInterface(wg.ServerParams{
		InfName:    WgInfName,
		Ip:         net.ParseIP(WgServerIp),
		Mask:       net.CIDRMask(WgPeerMask, 32),
		PrivateKey: conf.ServerPrivateKey,
		Port:       conf.ServerPort,
	}); err != nil {
		logrus.Fatal("Failed to add wg interface: ", err.Error())
	}

	logrus.Infof("Wg interface enabled")

	if err := tcTool.SetupTcBase(WgInfName, IfbInfName); err != nil {
		logrus.Fatal("Failed to add tc base: ", err.Error())
	}

	logrus.Infof("Tc base rules enabled")

	if err := iptablesTool.AddWgNatRules(ipt.NatParams{
		WgInf:  WgInfName,
		GwInf:  GwInfName,
		WgPort: conf.ServerPort,
		WgNet:  netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask)),
	}); err != nil {
		logrus.Fatal("Failed to add wg nat rules: ", err)
	}
	logrus.Infof("Nat rules enabled")

	//
	// 		Peer actions
	//

	for i, peer := range conf.Peers {
		if err := wgTool.AddWgPeer(wg.PeerParams{
			InfName:   WgInfName,
			Ip:        net.ParseIP(peer.Ip),
			Mask:      net.CIDRMask(WgPeerMask, 32),
			PublicKey: conf.PeerPublicKey,
		}); err != nil {
			logrus.Fatal("Failed to add wg peer: ", err.Error())
		}

		logrus.Infof("Wg peer %d enabled", i)

		if err := tcTool.ApplyTcForPeer(tc.PeerParams{
			WgInf:             WgInfName,
			IfbInf:            IfbInfName,
			PeerIp:            net.ParseIP(peer.Ip),
			ServerNetworkMask: net.CIDRMask(WgServerMask, 32),
			DownloadSpeedMb:   peer.Speed,
			UploadSpeedMb:     peer.Speed,
		}); err != nil {
			logrus.Fatal("Failed to add tc rules for peer: ", err.Error())
		}

		logrus.Infof("Tc rules for peer %d enabled", i)

	}

	go StartTimedValidator(context.Background(), 30*time.Second, wgTool, iptablesTool, tcTool, conf)

	// Wait for exit  signal
	waitForExitSignal()

	logrus.Info("Shutting down")

}

func StartTimedValidator(ctx context.Context, interval time.Duration, wgService wg.Tool, iptablesTool ipt.Tool, tcTool tc.Tool, config Configuration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			startTime := time.Now()
			err := Validate(wgService, iptablesTool, tcTool, config)
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

func Validate(wgService wg.Tool, iptablesTool ipt.Tool, tcTool tc.Tool, config Configuration) error {
	wgNetPrefix := netip.MustParsePrefix(fmt.Sprintf("%s/%d", WgServerIp, WgServerMask))

	if err := wgService.CheckWgInterface(wg.ServerParams{
		InfName:    WgInfName,
		Ip:         net.ParseIP(WgServerIp),
		Mask:       net.CIDRMask(WgServerMask, 32),
		PrivateKey: config.ServerPrivateKey,
		Port:       config.ServerPort,
	}); err != nil {
		return fmt.Errorf("wg interface check failed: %w", err)
	}

	if err := iptablesTool.CheckWgNatRules(ipt.NatParams{
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

	for i, peer := range config.Peers {
		if err := wgService.CheckWgPeer(wg.PeerParams{
			InfName:   WgInfName,
			Ip:        net.ParseIP(peer.Ip),
			Mask:      net.CIDRMask(WgPeerMask, 32),
			PublicKey: config.PeerPublicKey,
		}); err != nil {
			return fmt.Errorf("wg peer %d check failed: %w", i, err)
		}
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
