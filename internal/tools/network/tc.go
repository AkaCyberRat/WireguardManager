package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"WireguardManager/pkg/shell"

	"github.com/sirupsen/logrus"
)

func (t *Tool) tcServerUp() error {
	//
	// Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc add dev %s root handle 1: htb", t.interfaceName))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc add dev %s ingress", t.interfaceName))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for server enabled.")
	return nil
}

func (t *Tool) tcServerDown() error {
	//
	// Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc del dev %s root handle 1: htb", t.interfaceName))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc del dev %s ingress", t.interfaceName))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for server disabled.")
	return nil
}

func (t *Tool) tcPeerUp(ip string, downloadSpeed int, uploadSpeed int) error {
	ind := getIpIndex(ip)

	//
	// Limit download bandwidth
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc class add dev %s parent 1: classid 1:%v htb rate %vmbit ceil %vmbit", t.interfaceName, ind, downloadSpeed, downloadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip parent 1: prio %v u32 match ip src %v flowid 1:%v", t.interfaceName, ind, ip, ind))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip parent 1: prio %v u32 match ip dst %v flowid 1:%v", t.interfaceName, ind, ip, ind))
	if err != nil {
		return err
	}

	//
	// Limit upload bandwidth
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip src %v action police rate %vmbit burst 5mbit", t.interfaceName, ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip dst %v action police rate %vmbit burst 5mbit", t.interfaceName, ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for peer enabled. [Ip=%v]", ip)
	return nil
}

func (t *Tool) tcPeerDown(ip string) error {
	ind := getIpIndex(ip)

	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc filter del dev %s parent 1: prio %v", t.interfaceName, ind))
	if err != nil {
		return err
	}
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter del dev %s ingress prio %v", t.interfaceName, ind))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc class del dev %s parent 1: classid 1:%v", t.interfaceName, ind))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for peer disabled. [Ip=%v]", ip)
	return nil
}

func SetupTcBase(wgInf string) error {
	//
	// Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc add dev %s root handle 1: htb", wgInf))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc qdisc add dev %s ingress", wgInf))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for server enabled.")
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
	hostNum := hostNumber(peerIp, serverNetworkMask)

	// Limit download bandwidth (for server egress/upload)

	_, err := shell.RunExecWithTimeout(fmt.Sprintf("tc class add dev %[1]s parent 1: classid 1:%[2]v htb rate %[3]vmbit ceil %[3]vmbit", wgInf, hostNum, downloadSpeedMb))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %[1]s protocol ip parent 1: prio %[2]v u32 match ip src %[3]v flowid 1:%[2]v", wgInf, hostNum, peerIp))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %[1]s protocol ip parent 1: prio %[2]v u32 match ip dst %[3]v flowid 1:%[2]v", wgInf, hostNum, peerIp))
	if err != nil {
		return err
	}

	// Limit upload bandwidth (for server ingress/download)

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip src %v action police rate %vmbit burst 5mbit", wgInf, hostNum, peerIp, uploadSpeedMb))
	if err != nil {
		return err
	}

	_, err = shell.RunExecWithTimeout(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip dst %v action police rate %vmbit burst 5mbit", wgInf, hostNum, peerIp, uploadSpeedMb))
	if err != nil {
		return err
	}

	return nil
}

// DiscardTcForPeer removes the traffic control rules for a peer with the given IP address.
// It calculates the host number based on the peer's IP and the server's network mask, and then uses that host number to delete the tc rules for that peer.
//
// - 'wgInf' is the name of the Wireguard interface (e.g. 'wg0')
//
// - 'peerIp' is the IP address of the peer (e.g. '11.0.0.1')
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

func getIpIndex(ip string) int {
	octs := strings.Split(ip, ".")
	oct3, _ := strconv.Atoi(octs[2])
	oct4, _ := strconv.Atoi(octs[3])

	return (oct3 << 8) | oct4
}
