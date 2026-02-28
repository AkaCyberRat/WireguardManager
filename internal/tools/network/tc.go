package network

import (
	"WireguardManager/pkg/shell"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"net"
)

// SetupTcBase sets up the base traffic control rules for the Wireguard interface and IFB interface.
//
// - 'wgInf' is the name of the Wireguard interface (e.g. 'wg0')
//
// - 'ifbInf' is the name of the IFB interface to be used for ingress shaping (e.g. 'ifb0')
func SetupTcBase(wgInf, ifbInf string) error {
	//
	// 		Add base rule(s) to limit server egress (client download bandwidth)
	//

	// Create root HTB qdisc for wg interface
	// Command analog: 'tc qdisc add dev wgInf root handle 1: htb'
	wgInfLink, err := netlink.LinkByName(wgInf)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(wgInfLink); err != nil {
		return err
	}

	htbQdisc := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: wgInfLink.Attrs().Index,
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
	// Command analog: 'ip link add ifbInf type ifb'
	ifb := &netlink.Ifb{
		LinkAttrs: netlink.LinkAttrs{
			Name: ifbInf,
		},
	}

	if err := netlink.LinkAdd(ifb); err != nil {
		return err
	}

	// Command analog: 'ip link set ifbInf up'
	ifbInfLink, err := netlink.LinkByName(ifbInf)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(ifbInfLink); err != nil {
		return err
	}

	// Create ingress qdisc and filter on wg interface and redirect all ingress traffic to IFB interface
	// Command analog: 'tc filter add dev wgInf parent ffff: protocol ip u32 match ip src 0.0.0.0/0 action mirred egress redirect dev ifbInf'
	ingressQdisc := &netlink.Ingress{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: wgInfLink.Attrs().Index,
			Handle:    netlink.MakeHandle(0xffff, 0),
			Parent:    netlink.HANDLE_INGRESS,
		},
	}

	if err = netlink.QdiscAdd(ingressQdisc); err != nil {
		return err
	}

	// Command analog: 'tc qdisc add dev wgInf handle ffff: ingress'
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

	// Create root HTB qdisc for IFB interface to shape ingress traffic
	// Command analog: 'tc qdisc add dev ifbInf root handle 1: htb default 999'
	htb := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: ifbInfLink.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	})
	htb.Defcls = 999

	if err = netlink.QdiscAdd(htb); err != nil {
		return err
	}

	// Create default class with very high rate to avoid shaping traffic without specific rules
	// Command analog: 'tc class add dev ifbInf parent 1: classid 1:1 htb rate 100gbit'
	classAttrs := netlink.ClassAttrs{
		LinkIndex: ifbInfLink.Attrs().Index,
		Parent:    netlink.MakeHandle(1, 0),
		Handle:    netlink.MakeHandle(1, 1),
	}

	htbClass := netlink.NewHtbClass(classAttrs, netlink.HtbClassAttrs{
		Rate: 100 * 1000 * 1000 * 1000 / 8, // bytes per second! - 100gbit
	})

	if err = netlink.ClassAdd(htbClass); err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for server enabled.")
	return nil
}

// CheckTcBaseRules checks the existence of the base traffic control rules for the Wireguard interface and IFB interface.
//
// - 'wgInf' is the name of the Wireguard interface (e.g. 'wg0')
//
// - 'ifbInf' is the name of the IFB interface to be used for ingress shaping (e.g. 'ifb0')
//
// Returns an error if any of the required rules are missing or if there is an issue accessing the network interfaces.
func CheckTcBaseRules(wgInf, ifbInf string) error {
	// Check creation of WG interface
	wgInfLink, err := tryLink(wgInf)
	if err != nil {
		return err
	}

	//	Check root HTB qdisc for wg interface
	if err = hasQdisc(wgInfLink, netlink.MakeHandle(1, 0), netlink.HANDLE_ROOT); err != nil {
		return fmt.Errorf("wg root htb qdisc check failed: %s", err)
	}

	// Check IFB interface
	ifbInfLink, err := tryLink(ifbInf)
	if err != nil {
		return err
	}

	// Check IFB interface up
	if ifbInfLink.Attrs().Flags&net.FlagUp == 0 {
		return errors.New("IFB interface is not up")
	}

	// Check ingress qdisc on wg interface
	if err = hasQdisc(wgInfLink, netlink.MakeHandle(0xffff, 0), netlink.HANDLE_INGRESS); err != nil {
		return fmt.Errorf("wg ingress qdisc check failed: %s", err)
	}

	// Check creation of filter on wg interface and redirect all ingress traffic to IFB interface
	if err = hasU32EgressRedirectFilter(wgInfLink, netlink.MakeHandle(0xffff, 0)); err != nil {
		return fmt.Errorf("wg ingress redirect filter check failed: %s", err)
	}

	// Check creation of root HTB qdisc for IFB interface to shape ingress traffic
	if err = hasQdisc(ifbInfLink, netlink.MakeHandle(1, 0), netlink.HANDLE_ROOT); err != nil {
		return fmt.Errorf("ifb root htb qdisc check failed: %s", err)
	}

	// Check creation of default class with very high rate to avoid shaping traffic without specific rules
	if err = hasClass(ifbInfLink, netlink.MakeHandle(1, 1), netlink.MakeHandle(1, 0)); err != nil {
		return fmt.Errorf("ifb default htb class check failed: %s", err)
	}

	return nil
}

// ApplyTcForPeer applies traffic control rules for a peer with the given IP address, download speed, and upload speed.
// It calculates the host number based on the peer's IP and the server's network mask, and then uses that host number to create unique tc rules for that peer.
//
// - 'wgInf' is the name of the Wireguard interface (e.g. 'wg0')
//
// - 'ifbInf' is the name of the IFB interface to be used for ingress shaping (e.g. 'ifb0')
//
// - 'peerIp' is the IP address of the peer (e.g. '11.0.0.1')
//
// - 'serverNetworkMask' is the subnet mask of the server's network (e.g. '24')
//
// - 'downloadSpeedMb' is the download speed limit for the peer in megabits per second (e.g. 100)
//
// - 'uploadSpeedMb' is the upload speed limit for the peer in megabits per second (e.g. 50)
func ApplyTcForPeer(wgInf, ifbInf string, peerIp net.IP, serverNetworkMask net.IPMask, downloadSpeedMb, uploadSpeedMb uint) error {
	hostNum := hostNumber(peerIp, serverNetworkMask)

	//
	// 		Add rules to limit server egress for peer (client download bandwidth)
	//

	// Create class for peer with specified download speed
	// Command analog: 'tc class add dev wgInf parent 1:1 classid 1:{hostNum} htb rate {RATE}mbit ceil {RATE}mbit'
	wgInfLink, err := tryLink(wgInf)
	if err != nil {
		return err
	}

	classID := netlink.MakeHandle(1, uint16(hostNum))
	parent := netlink.MakeHandle(1, 1)
	rate := uint64(downloadSpeedMb) * 1000 * 1000 // mbit → bytes/sec
	ceil := rate

	classAttrs := netlink.ClassAttrs{
		LinkIndex: wgInfLink.Attrs().Index,
		Parent:    parent,
		Handle:    classID,
	}
	htbAttrs := netlink.HtbClassAttrs{
		Rate: rate,
		Ceil: ceil,
	}
	class := netlink.NewHtbClass(classAttrs, htbAttrs)
	if err = netlink.ClassAdd(class); err != nil {
		return err
	}

	// Create filters to direct traffic to peer to the class
	// 'tc filter add dev wgInf protocol ip parent 1: prio {hostNum} u32 match ip src {peerIp} flowid 1:{hostNum}'
	// 'tc filter add dev wgInf protocol ip parent 1: prio {hostNum} u32 match ip dst {peerIp} flowid 1:{hostNum}'

	commands := []string{

		// Create filters to direct traffic to peer to the class
		fmt.Sprintf("tc filter add dev %[1]s protocol ip parent 1: prio %[2]v u32 match ip src %[3]v flowid 1:%[2]v", wgInf, hostNum, peerIp),
		//fmt.Sprintf("tc filter add dev %[1]s protocol ip parent 1: prio %[2]v u32 match ip dst %[3]v flowid 1:%[2]v", wgInf, hostNum, peerIp),

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

func tryLink(name string) (netlink.Link, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		if errors.As(err, new(netlink.LinkNotFoundError)) {
			return nil, fmt.Errorf("%s interface doesn't exist: %w", name, err)
		}
		return nil, fmt.Errorf("failed to check %s interface: %w", name, err)
	}
	return link, nil
}

func hasQdisc(link netlink.Link, handle, parent uint32) error {
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return err
	}

	for _, q := range qdiscs {
		attrs := q.Attrs()
		if attrs.Handle == handle && attrs.Parent == parent {
			return nil
		}
	}
	return fmt.Errorf("qdisc not found ( Interface: '%s', Handle: '%v', Parent: '%v' )", link.Attrs().Name, handle, parent)
}

func hasClass(link netlink.Link, handle, parent uint32) error {
	classes, err := netlink.ClassList(link, parent)
	if err != nil {
		return err
	}

	for _, c := range classes {
		if c.Attrs().Handle == handle {
			return nil
		}
	}
	return fmt.Errorf("class not found ( Interface: '%s', Handle: '%v', Parent: '%v' )", link.Attrs().Name, handle, parent)
}

func hasU32EgressRedirectFilter(link netlink.Link, parent uint32) error {
	filters, err := netlink.FilterList(link, parent)
	if err != nil {
		return err
	}

	for _, filter := range filters {
		if filter, ok := filter.(*netlink.U32); ok {
			for _, act := range filter.Actions {
				if mirred, ok := act.(*netlink.MirredAction); ok {
					if mirred.MirredAction == netlink.TCA_EGRESS_REDIR &&
						mirred.Action == netlink.TC_ACT_STOLEN {
						return nil
					}
				}
			}
		}
	}
	return fmt.Errorf("redirect filter not found ( Interface: '%s', Parent: '%v' )", link.Attrs().Name, parent)
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
