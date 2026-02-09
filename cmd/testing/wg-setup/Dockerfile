FROM ubuntu:22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    iptables \
    wireguard \
    iproute2 \
    nftables && \
    rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy


WORKDIR /app

# бинарник будет подмонтирован volume-ом
ENTRYPOINT ["./app", " --config config.json"]
