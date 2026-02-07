#!/bin/bash
set -e

iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -i wg0 -j ACCEPT;
/coredns -conf /routing/corefile.conf &

exec "$@"