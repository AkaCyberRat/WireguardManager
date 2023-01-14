package network

import (
	"fmt"
	"strconv"
	"strings"

	"WireguardManager/pkg/shell"

	"github.com/sirupsen/logrus"
)

func (t *Tool) tcServerUp() error {
	//
	// Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.Run(fmt.Sprintf("tc qdisc add dev %s root handle 1: htb", t.interfaceName))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.Run(fmt.Sprintf("tc qdisc add dev %s ingress", t.interfaceName))
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
	_, err := shell.Run(fmt.Sprintf("tc qdisc del dev %s root handle 1: htb", t.interfaceName))
	if err != nil {
		return err
	}

	//
	// Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.Run(fmt.Sprintf("tc qdisc del dev %s ingress", t.interfaceName))
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
	_, err := shell.Run(fmt.Sprintf("tc class add dev %s parent 1: classid 1:%v htb rate %vmbit ceil %vmbit", t.interfaceName, ind, downloadSpeed, downloadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc filter add dev %s protocol ip parent 1: prio %v u32 match ip src %v flowid 1:%v", t.interfaceName, ind, ip, ind))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc filter add dev %s protocol ip parent 1: prio %v u32 match ip dst %v flowid 1:%v", t.interfaceName, ind, ip, ind))
	if err != nil {
		return err
	}

	//
	// Limit upload bandwidth
	//
	_, err = shell.Run(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip src %v action police rate %vmbit burst 5mbit", t.interfaceName, ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc filter add dev %s protocol ip ingress prio %v u32 match ip dst %v action police rate %vmbit burst 5mbit", t.interfaceName, ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for peer enabled. [Ip=%v]", ip)
	return nil
}

func (t *Tool) tcPeerDown(ip string) error {
	ind := getIpIndex(ip)

	_, err := shell.Run(fmt.Sprintf("tc filter del dev %s parent 1: prio %v", t.interfaceName, ind))
	if err != nil {
		return err
	}
	_, err = shell.Run(fmt.Sprintf("tc filter del dev %s ingress prio %v", t.interfaceName, ind))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc class del dev %s parent 1: classid 1:%v", t.interfaceName, ind))
	if err != nil {
		return err
	}

	logrus.Tracef("Traffic control rules for peer disabled. [Ip=%v]", ip)
	return nil
}

func getIpIndex(ip string) int {
	octs := strings.Split(ip, ".")
	oct3, _ := strconv.Atoi(octs[2])
	oct4, _ := strconv.Atoi(octs[3])

	return (oct3 << 8) | oct4
}
