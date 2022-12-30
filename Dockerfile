FROM golang:buster

RUN \ 
echo 'deb http://ftp.debian.org/debian buster-backports main' | tee /etc/apt/sources.list.d/buster-backports.list &&\
apt-get update &&\
apt-get install sudo -y wireguard dkms git gnupg ifupdown iproute2 iptables iputils-ping jq libc6 libelf-dev net-tools openresolv systemctl &&\
\
mkdir /app &&\
mkdir /app/src &&\
mkdir /app/config &&\
mkdir /app/additional &&\
\
echo "**** install CoreDNS ****" && \
COREDNS_VERSION=$(curl -sX GET "https://api.github.com/repos/coredns/coredns/releases/latest" | awk '/tag_name/{print $4;exit}' FS='[""]' | awk '{print substr($1,2); }') && \
curl -o /tmp/coredns.tar.gz -L "https://github.com/coredns/coredns/releases/download/v${COREDNS_VERSION}/coredns_${COREDNS_VERSION}_linux_amd64.tgz" && \
tar xf /tmp/coredns.tar.gz -C /app && \
\
echo "**** clean up ****" && \
rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

WORKDIR /app
COPY . /src

RUN \
 mv /src/additional /app/additional &&\
 mv /src/additional/coredns.service /etc/systemd/system/ &&\
 systemctl daemon-reload &&\
 chmod 777 additional/run.sh &&\
 cd /app/src &&\
 go build -o /app/main main.go 

EXPOSE 51830/udp
EXPOSE 5000
CMD [ "/app/additional/SetupIptableAndRunApp.sh" ]
#ENTRYPOINT ["./main"]