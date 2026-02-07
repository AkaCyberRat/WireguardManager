package main

import (
	"WireguardManager/internal/core"
	"flag"
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"WireguardManager/internal/config"
	"WireguardManager/internal/logging"
	"WireguardManager/internal/tools/network"
	"github.com/google/nftables"
	"github.com/google/nftables/expr"
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

func setupNATRouteAndMSS(wgIface, extIface string) error {
	// ============================
	// 1. NAT (nftables)
	// ============================
	c := &nftables.Conn{}

	nat := &nftables.Table{
		Name:   "nat",
		Family: nftables.TableFamilyIPv4,
	}
	c.AddTable(nat)

	postrouting := &nftables.Chain{
		Name:     "postrouting",
		Table:    nat,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	}
	c.AddChain(postrouting)

	c.AddRule(&nftables.Rule{
		Table: nat,
		Chain: postrouting,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyOIFNAME, Register: 1},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     append([]byte(extIface), 0),
			},
			&expr.Masq{},
		},
	})

	if err := c.Flush(); err != nil {
		return fmt.Errorf("nftables NAT failed: %w", err)
	}

	// ============================
	// 2. Route via wg
	// ============================
	wg, err := netlink.LinkByName(wgIface)
	if err != nil {
		return err
	}

	route := &netlink.Route{
		LinkIndex: wg.Attrs().Index,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 0)},
	}
	_ = netlink.RouteReplace(route)

	// ============================
	// 3. MSS clamping (iptables)
	// ============================
	cmd := exec.Command(
		"iptables",
		"-t", "mangle",
		"-C", "FORWARD",
		"-p", "tcp",
		"--tcp-flags", "SYN,RST", "SYN",
		"-j", "TCPMSS",
		"--clamp-mss-to-pmtu",
	)

	if err := cmd.Run(); err != nil {

		// правила нет — добавляем
		err = exec.Command(
			"iptables",
			"-t", "mangle",
			"-A", "FORWARD",
			"-p", "tcp",
			"--tcp-flags", "SYN,RST", "SYN",
			"-j", "TCPMSS",
			"--clamp-mss-to-pmtu",
		).Run()
		if err != nil {
			logrus.Fatal(err)
		}
	}

	return nil
}

//// setupNATAndRoute настраивает NAT (MASQUERADE) на extIface
//// и добавляет маршрут 0.0.0.0/0 через wgIface.
//func setupNATAndRoute(wgIface, extIface string) error {
//	// ----------------------------
//	// 1️⃣ Настройка NAT через nftables
//	// ----------------------------
//	c := &nftables.Conn{}
//
//	table := &nftables.Table{
//		Name:   "nat",
//		Family: nftables.TableFamilyIPv4,
//	}
//	c.AddTable(table)
//
//	chain := &nftables.Chain{
//		Name:     "POSTROUTING",
//		Table:    table,
//		Type:     nftables.ChainTypeNAT,
//		Hooknum:  nftables.ChainHookPostrouting,
//		Priority: nftables.ChainPriorityNATSource,
//	}
//	c.AddChain(chain)
//
//	// MASQUERADE для пакетов исходящих через extIface
//	c.AddRule(&nftables.Rule{
//		Table: table,
//		Chain: chain,
//		Exprs: []expr.Any{
//			// Сравниваем имя исходящего интерфейса
//			&expr.Meta{Key: expr.MetaKeyOIFNAME, Register: 1},
//			&expr.Cmp{
//				Op:       expr.CmpOpEq,
//				Register: 1,
//				Data:     append([]byte(extIface), 0), // null-terminated
//			},
//			&expr.Masq{},
//		},
//	})
//
//	if err := c.Flush(); err != nil {
//		return fmt.Errorf("failed to flush nftables rules: %v", err)
//	}
//
//	// ----------------------------
//	// 2️⃣ Настройка маршрута 0.0.0.0/0 через wgIface
//	// ----------------------------
//	link, err := netlink.LinkByName(wgIface)
//	if err != nil {
//		return fmt.Errorf("failed to find wg interface: %v", err)
//	}
//
//	defaultRoute := &netlink.Route{
//		LinkIndex: link.Attrs().Index,
//		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 0)},
//	}
//
//	// Добавляем маршрут
//	if err := netlink.RouteAdd(defaultRoute); err != nil {
//		// Если маршрут уже есть, игнорируем ошибку
//		if !os.IsExist(err) {
//			return fmt.Errorf("failed to add default route: %v", err)
//		}
//	}
//
//	return nil
//}

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

	wgIface := "wg0"   // Ваш WireGuard интерфейс
	extIface := "eth0" // Внешний интерфейс

	if err := setupNATRouteAndMSS(wgIface, extIface); err != nil {
		log.Fatal("Failed to setup NAT and route: ", err)
	}

	log.Println("NAT и маршруты успешно настроены")

	err = netTool.EnablePeer(&core.Peer{
		Id:            "2",
		Ip:            "10.0.0.2",
		PublicKey:     conf.PeerPublicKey,
		PresharedKey:  "",
		DownloadSpeed: 10,
		UploadSpeed:   10,
		Status:        core.Enabled,
	})
	if err != nil {
		logrus.Fatal("Failed to enable peer:", err.Error())
	}

	logrus.Infof("Peer enabled")

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
