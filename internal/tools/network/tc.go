package network

import (
	"WireguardManager/pkg/shell"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func SetupTcBase(wgInf string) error {
	const ifbInf = "ifb0"
	const rootRate = "100gbit"

	//
	// 		Add base rule(s) to limit server egress (client download bandwidth)
	//

	// Create root HTB qdisc for wg interface
	// tc qdisc add dev wgInf root handle 1: htb
	link, err := netlink.LinkByName(wgInf)
	if err != nil {
		fmt.Println("LinkByName error:", err)
		return err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	htbQdisc := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	})

	if err := netlink.QdiscAdd(htbQdisc); err != nil {
		return err
	}

	//
	// 		Add base rule(s) to limit server ingress (client upload bandwidth)
	//

	// Create and up IFB interface
	// "ip link add ifbInf type ifb"
	// "ip link set ifbInf up"

	ifb := &netlink.Ifb{
		LinkAttrs: netlink.LinkAttrs{
			Name: ifbInf,
		},
	}

	if err := netlink.LinkAdd(ifb); err != nil {
		return err
	}

	link, err = netlink.LinkByName(ifbInf)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	// Create ingress qdisc and filter on wg interface and redirect all ingress traffic to IFB interface
	// "tc qdisc add dev wgInf handle ffff: ingress"
	// "tc filter add dev wgInf parent ffff: protocol ip u32 match ip src 0.0.0.0/0 action mirred egress redirect dev ifbInf"

	link, err = netlink.LinkByName(wgInf)
	if err != nil {
		return err
	}

	ingressQdisc := &netlink.Ingress{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(0xffff, 0),
			Parent:    netlink.HANDLE_INGRESS,
		},
	}

	if err = netlink.QdiscAdd(ingressQdisc); err != nil {
		return err
	}

	wgInfLink, err := netlink.LinkByName(wgInf)
	if err != nil {
		return err
	}

	ifbInfLink, err := netlink.LinkByName(ifbInf)
	if err != nil {
		return err
	}

	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: wgInfLink.Attrs().Index,
			Parent:    netlink.MakeHandle(0xffff, 0),
			Protocol:  unix.ETH_P_IP,
			Priority:  1,
		},
		Actions: []netlink.Action{
			&netlink.MirredAction{
				ActionAttrs: netlink.ActionAttrs{
					Action: netlink.TC_ACT_STOLEN,
				},
				MirredAction: netlink.TCA_EGRESS_REDIR,
				Ifindex:      ifbInfLink.Attrs().Index,
			},
		},
	}

	if err = netlink.FilterAdd(filter); err != nil {
		return err
	}

	commands := []string{
		// Create root HTB qdisc for IFB interface to shape ingress traffic
		fmt.Sprintf("tc qdisc add dev %s root handle 1: htb default 999", ifbInf),

		// Create default class with very high rate to avoid shaping traffic without specific rules
		fmt.Sprintf("tc class add dev %s parent 1: classid 1:1 htb rate %s", ifbInf, rootRate),
	}

	for _, cmd := range commands {
		if _, err := shell.RunExecWithTimeout(cmd); err != nil {
			return err
		}
	}

	logrus.Tracef("Traffic control rules for server enabled.")
	return nil
}

func CheckTcBase(wgInf string) error {
	const ifbInf = "ifb0"

	//	Check creation of root HTB qdisc for wg interface
	link, err := netlink.LinkByName(wgInf)
	if err != nil {
		return err
	}

	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return err
	}

	for _, q := range qdiscs {
		if _, ok := q.(*netlink.Htb); !ok {
			return errors.New("Qdisc is not a Htb")
		}
	}

	// Check creation of IFB interface
	link, err = netlink.LinkByName(ifbInf)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return errors.New("IFB interface doesnt exist:" + err.Error())
		}

		return errors.New("Failed to check IFB interface existence:" + err.Error())
	}

	// Check if IFB interface is up
	if link.Attrs().Flags&net.FlagUp == 0 {
		return errors.New("IFB interface is not up")
	}

	// Check creation of ingress qdisc and filter on wg interface and redirect all ingress traffic to IFB interface

	link, err = netlink.LinkByName(wgInf)
	if err != nil {
		return err
	}

	qdiscs, err = netlink.QdiscList(link)
	if err != nil {
		return err
	}

	for _, q := range qdiscs {
		if _, ok := q.(*netlink.Ingress); !ok {
			return errors.New("qdisc is not a Ingress qdisc")
		}
	}

	filters, err := netlink.FilterList(link, netlink.MakeHandle(0xffff, 0))
	if err != nil {
		return err
	}

	if len(filters) == 0 {
		return errors.New("no redirect filters found on wg interface")
	}

	return nil
}

// ApplyTcForPeer applies traffic control rules for a peer with the given IP address, download speed, and upload speed.
// It calculates the host number based on the peer's IP and the server's network mask, and then uses that host number to create unique tc rules for that peer.
//
// - 'wgInf' is the name of the Wireguard interface (e.g. 'wg0')
//
// - 'peerIp' is the IP address of the peer (e.g. '11.0.0.1')
//
// - 'serverNetworkMask' is the subnet mask of the server's network (e.g. '24')
//
// - 'downloadSpeedMb' is the download speed limit for the peer in megabits per second (e.g. 100)
//
// - 'uploadSpeedMb' is the upload speed limit for the peer in megabits per second (e.g. 50)
func ApplyTcForPeer(wgInf string, peerIp net.IP, serverNetworkMask net.IPMask, downloadSpeedMb int, uploadSpeedMb int) error {
	const ifbInf = "ifb0"
	hostNum := hostNumber(peerIp, serverNetworkMask)

	commands := []string{
		//
		// 		Add rules to limit server egress for peer (client download bandwidth)
		//

		// Create class for peer with specified download speed
		fmt.Sprintf("tc class add dev %[1]s parent 1:1 classid 1:%[2]v htb rate %[3]vmbit ceil %[3]vmbit", wgInf, hostNum, downloadSpeedMb),

		// Create filters to direct traffic to peer to the class
		fmt.Sprintf("tc filter add dev %[1]s protocol ip parent 1: prio %[2]v u32 match ip src %[3]v flowid 1:%[2]v", wgInf, hostNum, peerIp),
		fmt.Sprintf("tc filter add dev %[1]s protocol ip parent 1: prio %[2]v u32 match ip dst %[3]v flowid 1:%[2]v", wgInf, hostNum, peerIp),

		//
		// 		Add rules to limit server ingress for peer (client upload bandwidth)
		//

		// Create class for peer with specified upload speed
		fmt.Sprintf("tc class add dev %s parent 1: classid 1:%d htb rate %vmbit ceil %vmbit", ifbInf, hostNum, uploadSpeedMb, uploadSpeedMb),

		// Create filters to direct traffic from peer to the class
		fmt.Sprintf("tc filter add dev %s parent 1: protocol ip prio 1 u32 match ip src %s flowid 1:%d", ifbInf, peerIp, hostNum),
		fmt.Sprintf("tc qdisc add dev %s parent 1:%d fq_codel", ifbInf, hostNum),
	}

	// TODO: Add error context
	for _, cmd := range commands {
		if _, err := shell.RunExecWithTimeout(cmd); err != nil {
			return err
		}
	}

	return nil
}

// DiscardTcForPeer removes the traffic control rules for a peer with the given IP address.
// It calculates the host number based on the peer's IP and the server's network mask, and then uses that host number to delete the tc rules for that peer.
//
// - 'wgInf' is the name of the Wireguard interface (e.g. 'wg0')
//
// - 'peerIp' is the IP address of the peer (e.g. '11.0.0.1')
// TODO: Actualize
func DiscardTcForPeer(wgInf string, peerIp net.IP, serverNetworkMask net.IPMask) error {
	hostNum := hostNumber(peerIp, serverNetworkMask)

	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc filter del dev %s parent 1: prio %v", wgInf, hostNum))
	if err != nil {
		return err
	}
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter del dev %s ingress prio %v", wgInf, hostNum))
	if err != nil {
		return err
	}
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc class del dev %s parent 1: classid 1:%v", wgInf, hostNum))
	if err != nil {
		return err
	}

	// wg ingress redirect
	_, _ = shell.RunExecWithTimeout(
		fmt.Sprintf("tc filter del dev %s ingress prio %v", wgInf, hostNum),
	)

	// ifb
	_, _ = shell.RunExecWithTimeout(
		fmt.Sprintf("tc filter del dev ifb0 parent 2: prio %v", hostNum),
	)
	_, _ = shell.RunExecWithTimeout(
		fmt.Sprintf("tc class del dev ifb0 parent 2: classid 2:%v", hostNum),
	)

	return nil
}

func hostNumber(ip net.IP, mask net.IPMask) uint32 {
	ip4 := ip.To4()
	if ip4 == nil {
		panic("only IPv4 supported")
	}

	// Преобразуем маску в uint32
	maskUint := uint32(mask[0])<<24 | uint32(mask[1])<<16 | uint32(mask[2])<<8 | uint32(mask[3])
	// Преобразуем IP в uint32
	ipUint := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])

	// host number = ip & ^mask
	return ipUint & ^maskUint
}

//modprobe ifb
//ip link add ifb0 type ifb
//ip link set ifb0 up
//
//tc qdisc add dev wg0 handle ffff: ingress
//tc filter add dev wg0 parent ffff: protocol ip u32 match ip src PEER_IP action mirred egress redirect dev ifb0
//
//tc qdisc add dev ifb0 root handle 1: htb
//tc class add dev ifb0 parent 1: classid 1:1 htb rate 100mbit ceil 100mbit

func getIpIndex(ip string) int {
	octs := strings.Split(ip, ".")
	oct3, _ := strconv.Atoi(octs[2])
	oct4, _ := strconv.Atoi(octs[3])

	return (oct3 << 8) | oct4
}
