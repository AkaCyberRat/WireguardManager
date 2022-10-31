#!/bin/bash
echo "Setup ip tables..."
iptables -t nat -A POSTROUTING -o eth0+ -j MASQUERADE
iptables -A FORWARD -i wg0 -j ACCEPT;
systemctl start coredns
systemctl enable coredns
echo "Complete"
./main