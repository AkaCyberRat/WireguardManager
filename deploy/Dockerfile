FROM golang:buster AS base

RUN \ 
echo 'deb http://ftp.debian.org/debian buster-backports main' | tee /etc/apt/sources.list.d/buster-backports.list &&\
apt-get update &&\
apt-get install sudo -y wireguard dkms git gnupg ifupdown iproute2 iptables iputils-ping jq libc6 libelf-dev net-tools openresolv systemctl &&\
\
echo "**** install CoreDNS ****" && \
COREDNS_VERSION=$(curl -sX GET "https://api.github.com/repos/coredns/coredns/releases/latest" | awk '/tag_name/{print $4;exit}' FS='[""]' | awk '{print substr($1,2); }') && \
curl -o /tmp/coredns.tar.gz -L "https://github.com/coredns/coredns/releases/download/v${COREDNS_VERSION}/coredns_${COREDNS_VERSION}_linux_amd64.tgz" && \
tar xf /tmp/coredns.tar.gz -C / && \
\
echo "**** clean up ****" && \
rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*



FROM golang:buster AS build

COPY go.* /src/
WORKDIR /src
RUN go mod download

COPY . /src
RUN \
mkdir /app && \
mkdir /app/deploy && \
mkdir /app/configs && \
\
cp -a /src/deploy/. /app/deploy && \
\
cd /src/cmd/api && \
go build -o /app/main main.go 



FROM base AS app

COPY --from=build /app/. /app
RUN \
cp /app/deploy/coredns.service /etc/systemd/system/coredns.service && \
systemctl daemon-reload && \
chmod 777 /app/deploy/run.sh

EXPOSE 51830/udp
EXPOSE 5000

CMD [ "/app/deploy/run.sh" ]

