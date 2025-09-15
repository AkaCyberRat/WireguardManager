ARG CMD_PATH=cmd/api
ARG CMD_NAME=main.go


FROM golang:buster AS base

RUN \
echo "**** Update packages ****" && \
echo 'deb http://ftp.debian.org/debian buster-backports main' | tee /etc/apt/sources.list.d/buster-backports.list &&\
apt-get update &&\
apt-get install sudo -y wireguard dkms git gnupg ifupdown iproute2 iptables iputils-ping jq libc6 libelf-dev net-tools openresolv systemctl &&\
\
echo "**** Install CoreDNS ****" && \
COREDNS_VERSION=$(curl -sX GET "https://api.github.com/repos/coredns/coredns/releases/latest" | awk '/tag_name/{print $4;exit}' FS='[""]' | awk '{print substr($1,2); }') && \
curl -o /tmp/coredns.tar.gz -L "https://github.com/coredns/coredns/releases/download/v${COREDNS_VERSION}/coredns_${COREDNS_VERSION}_linux_amd64.tgz" && \
tar xf /tmp/coredns.tar.gz -C / && \
\
echo "**** Clean up ****" && \
rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

COPY ./corefile.conf /routing/corefile.conf
COPY ./coredns.service /etc/systemd/system/coredns.service
COPY ./setup.sh /routing/setup.sh

RUN chmod 755 /routing/setup.sh



FROM golang:buster AS build

COPY go.* /src/

WORKDIR /src

RUN go mod download

COPY . /src

RUN go build -o /app/app /src/$CMD_PATH$CMD_NAME



FROM base AS app

COPY --from=build /app/. /app

WORKDIR /app

ENTRYPOINT [ "/routing/setup.sh" ]

CMD [ "/app/app" ]

