package main

import (
	"net"

	"WireguardManager/src/logger"
	"WireguardManager/src/network"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {

	logger.Configure()

	client, _ := wgctrl.New()

	network.CreateWgInf()

	err := network.ConfigureWgInf(client)
	if err != nil {
		logrus.Fatal("Cant configure wg inf: ", err.Error())
	}

	err = network.SetWgPeers(client, GeneratePeers(client), false)
	if err != nil {
		logrus.Fatal("Cant set wg peers: ", err.Error())
	}

	err = network.UpWgInf()
	if err != nil {
		logrus.Fatal("Cant up wg inf: ", err.Error())
	}

	err = network.TcConfigure()
	if err != nil {
		logrus.Fatal("Cant configure tc tool: ", err.Error())
	}

	for i := 0; i < 1; i++ {
		err = network.TcAddRules("10.1.1.2/32", 10, 10, 1, 100)
		if err != nil {
			logrus.Fatal("Cant configure tc tool: ", err.Error())
		}
	}

	// network.Execute(time.Second, "tc", "qdisc", "add", "dev", "wg0", "root", "handle", "1:", "htb", "default", "30")
	// network.Execute(time.Second, "tc", "class", "add", "dev", "wg0", "parent", "1:", "classid", "1:1", "htb", "rate", "100mbit", "burst", "10mbit")

	// network.Execute(time.Second, "tc", "class", "add", "dev", "wg0", "parent", "1:1", "classid", "1:10", "htb", "rate", "5mbit", "burst", "1mbit")
	// network.Execute(time.Second, "tc", "class", "add", "dev", "wg0", "parent", "1:1", "classid", "1:20", "htb", "rate", "5mbit", "burst", "1mbit")

	// network.Execute(time.Second, "tc", "qdisc", "add", "dev", "wg0", "parent", "1:10", "handle", "10:", "sfq", "perturb", "10")
	// network.Execute(time.Second, "tc", "qdisc", "add", "dev", "wg0", "parent", "1:20", "handle", "20:", "sfq", "perturb", "10")

	// $TC qdisc add dev $interface root handle 1: htb default 30
	// $TC class add dev $interface parent 1: classid 1:1 htb rate $interface_speed burst 15k

	// $TC class add dev $interface parent 1:1 classid 1:10 htb rate $download_limit burst 15k
	// $TC class add dev $interface parent 1:1 classid 1:20 htb rate $upload_limit burst 15k

	// $TC qdisc add dev $interface parent 1:10 handle 10: sfq perturb 10
	// $TC qdisc add dev $interface parent 1:20 handle 20: sfq perturb 10

	// network.Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "parent", "1:", "prio", "1", "u32", "match", "ip", "dst", "10.1.1.2/32", "flowid", "1:10")
	// network.Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "parent", "1:", "prio", "1", "u32", "match", "ip", "src", "10.1.1.2/32", "flowid", "1:20")

	// TC filter add dev $interface protocol ip parent 1: prio 1 u32 match ip dst 0.0.0.0/0 flowid 1:10
	// TC filter add dev $interface protocol ip parent 1: prio 1 u32 match ip src 0.0.0.0/0 flowid 1:20

	// # If you want to limit the upload/download limit based on specific IP address
	// # you can comment the above catch-all filter and uncomment these:
	// #
	// # $FILTER match ip dst $ip/32 flowid 1:10
	// # $FILTER match ip src $ip/32 flowid 1:20

	select {}
}

func GeneratePeers(client *wgctrl.Client) *[]wgtypes.PeerConfig {

	prKey, _ := wgtypes.ParseKey("uKIkAl5agqGLoodeDAdtgZHh91vXck5z/mmxETx2dWs=")
	pubKey := prKey.PublicKey()
	ipAddress := "10.1.1.2/32"

	logrus.Infof("Client pk: %v", prKey.String())
	logrus.Infof("Client pubk: %v", prKey.PublicKey().String())

	var ipAddresses []net.IPNet
	_, ipnet, err := net.ParseCIDR(ipAddress)
	if err != nil {
		logrus.Fatal("Cant parse ip net address for peer")
	}
	ipAddresses = append(ipAddresses, *ipnet)

	peer := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		AllowedIPs: ipAddresses,
	}

	var peers []wgtypes.PeerConfig
	peers = append(peers, peer)

	return &peers
}
