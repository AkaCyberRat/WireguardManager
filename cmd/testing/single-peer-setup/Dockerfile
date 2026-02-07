FROM alpine:3.18

RUN apk add --no-cache \
    iptables \
    iproute2 \
    nftables

WORKDIR /app

# бинарник будет подмонтирован volume-ом
ENTRYPOINT ["./single-peer-setup", " --config config.json"]
