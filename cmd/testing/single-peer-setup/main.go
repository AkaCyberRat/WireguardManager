package main

import (
	"WireguardManager/internal/core"
	"flag"
	"fmt"
	"os"
	"os/exec"
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

// runExec выполняет команду и возвращает ошибку с выводом
func runExec(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %v\n%s", cmd, args, err, out)
	}
	return nil
}

// SetupWgNAT настраивает MASQUERADE и форвардинг для WireGuard контейнера
func SetupWgNAT() error {

	//iptables -t nat -A POSTROUTING -s 11.0.0.0/24 -o eth0 -j MASQUERADE;
	//iptables -A INPUT -p udp -m udp --dport 51820 -j ACCEPT;
	//iptables -A FORWARD -i wg0 -j ACCEPT;
	//iptables -A FORWARD -o wg0 -j ACCEPT

	if err := runExec("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "11.0.0.0/24", "-o", "eth0", "-j", "MASQUERADE"); err != nil {
		return err
	}

	if err := runExec("iptables", "-A", "INPUT", "-p", "udp", "-m", "udp", "--dport", "51820", "-j", "ACCEPT"); err != nil {
		return err
	}

	if err := runExec("iptables", "-A", "FORWARD", "-i", "wg0", "-j", "ACCEPT"); err != nil {
		return err
	}

	if err := runExec("iptables", "-A", "FORWARD", "-o", "wg0", "-j", "ACCEPT"); err != nil {
		return err
	}

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

	err = SetupWgNAT()
	if err != nil {
		logrus.Fatal("Failed to setup NAT: ", err.Error())
	}

	if err != nil {
		logrus.Fatal("Failed to setup nftables: ", err.Error())
	}
	logrus.Infof("nftables rules set up")

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
