package ipt

import (
	tool "WireguardManager/internal/tools"
	"fmt"
	"net/netip"

	"github.com/coreos/go-iptables/iptables"
	"github.com/google/shlex"
)

// Tool that allow to manage legacy iptables rules(currently for local NAT).
type IptablesTool struct {
	ipt *iptables.IPTables
}

func NewIptablesTool() (IptablesTool, error) {
	var empty IptablesTool

	ipt, err := iptables.New()
	if err != nil {
		return empty, err
	}

	return IptablesTool{ipt: ipt}, nil
}

// WgNatParams describes parameters for WireGuard NAT rules.
type WgNatParams struct {
	// WgInf is the name of WireGuard interface (e.g. wg0).
	WgInf string
	// GwInf is the name of gateway interface (e.g. eth0).
	GwInf string
	// WgPort is the port on which WireGuard server listens (e.g. 51820).
	WgPort int
	// WgNet is the CIDR prefix of WireGuard network (e.g. 11.0.0.0/24).
	WgNet netip.Prefix
}

// AddWgNatRules turns on ipt (legacy) NAT for packets forwarding between interfaces.
func (t *IptablesTool) AddWgNatRules(params WgNatParams) error {
	rules := wgNatRules(params)
	return t.addRules(rules)
}

// AddDnsWgNatRules turns on ipt (legacy) NAT for packets forwarding between interfaces.
//
// - 'wgInf' 	is the name of wg interface (e.g. wg0)
func (t *IptablesTool) AddDnsWgNatRules(wgInf string) error {
	rules := wgDnsNatRules(wgInf)
	return t.addRules(rules)
}

// CheckWgNatRules checks ipt (legacy) NAT rules existence.
// Returns nil if all rules exist, returns ErrNotExist if any rule does not exist, otherwise returns an error.
func (t *IptablesTool) CheckWgNatRules(params WgNatParams) error {
	rules := wgNatRules(params)
	return t.checkRules(rules)
}

// CheckWgDnsNatRules checks ipt (legacy) NAT rules existence.
//
// - 'wgInf' 	is the name of wg interface (e.g. wg0)
//
// Returns nil if all rules exist, returns ErrNotExist if any rule does not exist, otherwise returns an error.
func (t *IptablesTool) CheckWgDnsNatRules(wgInf string) error {
	rules := wgDnsNatRules(wgInf)
	return t.checkRules(rules)
}

// DeleteWgNat delete ipt (legacy) NAT rules if exists.
func (t *IptablesTool) DeleteWgNat(params WgNatParams) error {
	rules := wgNatRules(params)
	return t.deleteRules(rules)
}

// DeleteDnsWgNat delete ipt (legacy) NAT rules if exists.
//
// - 'wgInf' 	is the name of wg interface (e.g. wg0)
func (t *IptablesTool) DeleteDnsWgNat(wgInf string) error {
	rules := wgDnsNatRules(wgInf)
	return t.deleteRules(rules)
}

func (t *IptablesTool) addRules(rules []string) error {
	for _, rule := range rules {
		args, err := shlex.Split(rule)
		if err != nil {
			return err
		}

		if err = t.ipt.AppendUnique(args[1], args[3], args[4:]...); err != nil {
			return err
		}
	}

	return nil
}

func (t *IptablesTool) checkRules(rules []string) error {
	for _, rule := range rules {
		args, err := shlex.Split(rule)
		if err != nil {
			return err
		}

		exists, err := t.ipt.Exists(args[1], args[3], args[4:]...)
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("rule %w ( Rule: '%v')", tool.ErrNotExist, rule)
		}
	}

	return nil
}

func (t *IptablesTool) deleteRules(rules []string) error {
	for _, rule := range rules {
		args, err := shlex.Split(rule)
		if err != nil {
			return err
		}

		if err = t.ipt.DeleteIfExists(args[1], args[3], args[4:]...); err != nil {
			return err
		}
	}

	return nil
}

func wgNatRules(params WgNatParams) []string {
	return []string{
		// NAT for wg interface
		fmt.Sprintf("-t nat -A POSTROUTING -s %v -o %v -j MASQUERADE", params.WgNet, params.GwInf),
		fmt.Sprintf("-t filter -A INPUT -p udp -m udp --dport %v -j ACCEPT", params.WgPort),
		fmt.Sprintf("-t filter -A FORWARD -i %v -j ACCEPT", params.WgInf),
		fmt.Sprintf("-t filter -A FORWARD -o %v -j ACCEPT", params.WgInf),
	}
}

// TODO: Add switchable dns destination
func wgDnsNatRules(wgInf string) []string {
	return []string{
		// DNS redirect
		fmt.Sprintf("-t nat -A PREROUTING -i %v -p udp --dport 53 -j DNAT --to-destination 8.8.8.8:53", wgInf),
		fmt.Sprintf("-t nat -A PREROUTING -i %v -p tcp --dport 53 -j DNAT --to-destination 8.8.8.8:53", wgInf),
	}
}
