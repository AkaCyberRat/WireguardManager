package network

import (
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func CreateWgInf() error {
	isExists := isWgInfExists()
	if isExists {
		return nil
	}

	// handle, err := netlink.NewHandle()
	// if err != nil {
	// 	return err
	// }

	h, _ := netlink.NewHandle()

	linkAtrrs := netlink.NewLinkAttrs()
	linkAtrrs.Name = "wg0"
	linkAtrrs.MTU = 1420
	linkAtrrs.TxQLen = 1000

	link := wgLink{}
	link.LinkAttrs = &linkAtrrs
	link.LinkType = "wireguard"

	err := h.LinkAdd(netlink.Link(link))
	return err
}

func ConfigureWgInf(client *wgctrl.Client) error {

	port := 51830
	ipv4 := "10.1.1.1/16"
	privateKey, err := wgtypes.ParseKey("8EUc3jnJUrWoZVDFA4KL6Rs++a34sfMmukYWxnAGDmA=") //wgtypes.GeneratePrivateKey()
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &privateKey,
		ListenPort:   &port,
		ReplacePeers: false,
		Peers:        make([]wgtypes.PeerConfig, 0),
	}

	logrus.Infof("Server ip: %v", ipv4)
	logrus.Info("Server port: 51830")
	logrus.Infof("Server pk: %v", privateKey.String())
	logrus.Infof("Server pk: %v", privateKey.PublicKey().String())

	h, _ := netlink.NewHandle()

	link, _ := h.LinkByName("wg0")
	addr, _ := netlink.ParseAddr(ipv4)
	err = h.AddrAdd(link, addr)
	if err != nil {
		return err
	}

	err = client.ConfigureDevice("wg0", config)
	return err
}

func DeleteWgInf() error {
	// handle, err := netlink.NewHandle()
	// if err != nil {
	// 	return err
	// }

	linkAtrrs := netlink.NewLinkAttrs()
	linkAtrrs.Name = "wg0"
	linkAtrrs.MTU = 1420
	linkAtrrs.TxQLen = 1000

	link := wgLink{}
	link.LinkAttrs = &linkAtrrs
	link.LinkType = "wireguard"

	err := netlink.LinkDel(netlink.Link(link))
	return err
}

func UpWgInf() error {
	h, _ := netlink.NewHandle()

	link, err := h.LinkByName("wg0")
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(netlink.Link(link))
	return err
}

func DownWgInf() error {

	link, err := netlink.LinkByName("wg0")
	if err != nil {
		return err
	}

	err = netlink.LinkSetDown(netlink.Link(link))
	return err
}

func SetWgPeers(client *wgctrl.Client, peers *[]wgtypes.PeerConfig, replace bool) error {
	return client.ConfigureDevice("wg0", wgtypes.Config{Peers: *peers})
}
