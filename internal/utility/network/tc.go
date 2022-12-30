package network

import (
	"fmt"
	"strconv"
	"strings"

	"WireguardManager/pkg/shell"
)

func (t *Tool) tcConfigure() error {
	//
	//Add tc base rule to limit client download bandwidth (server upload)
	//
	_, err := shell.Run("tc qdisc add dev wg0 root handle 1: htb")
	if err != nil {
		return err
	}

	//
	//Add tc base rule to limit client upload bandwidth (server download)
	//
	_, err = shell.Run("tc qdisc add dev wg0 ingress")
	if err != nil {
		return err
	}

	return nil
}

func (t *Tool) tcPeerUp(ip string, downloadSpeed int, uploadSpeed int) error {
	ind := getIpIndex(ip)

	//
	// Limit download bandwidth
	//
	_, err := shell.Run(fmt.Sprintf("tc class add dev wg0 parent 1: classid 1:%v htb rate %vmbit ceil %vmbit", ind, downloadSpeed, downloadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc filter add dev wg0 protocol ip parent 1: prio %v u32 match ip src %v flowid 1:%v", ind, ip, ind))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc filter add dev wg0 protocol ip parent 1: prio %v u32 match ip dst %v flowid 1:%v", ind, ip, ind))
	if err != nil {
		return err
	}

	//
	// Limit upload bandwidth
	//
	_, err = shell.Run(fmt.Sprintf("tc filter add dev wg0 protocol ip ingress prio %v u32 match ip src %v action police rate %vmbit burst 5mbit", ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc filter add dev wg0 protocol ip ingress prio %v u32 match ip dst %v action police rate %vmbit burst 5mbit", ind, ip, uploadSpeed))
	if err != nil {
		return err
	}

	return nil
}

func (t *Tool) tcPeerDown(ip string) error {
	ind := getIpIndex(ip)

	_, err := shell.Run(fmt.Sprintf("tc filter del dev wg0 parent 1: prio %v", ind))
	if err != nil {
		return err
	}
	_, err = shell.Run(fmt.Sprintf("tc filter del dev wg0 ingress prio %v", ind))
	if err != nil {
		return err
	}

	_, err = shell.Run(fmt.Sprintf("tc class del dev wg0 parent 1: classid 1:%v", ind))
	if err != nil {
		return err
	}

	return nil
}

func getIpIndex(ip string) int {
	octs := strings.Split(ip, ".")
	oct3, _ := strconv.Atoi(octs[2])
	oct4, _ := strconv.Atoi(octs[3])

	return (oct3 << 8) | oct4
}
