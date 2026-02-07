FROM golang:1.23-bullseye AS base

RUN echo "**** Update packages ****" && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        sudo \
        wireguard \
        dkms \
        git \
        gnupg \
        ifupdown \
        iproute2 \
        iptables \
        iputils-ping \
        jq \
        libc6 \
        libelf-dev \
        net-tools \
        openresolv \
        systemd

RUN echo "**** Install CoreDNS ****" && \
    COREDNS_VERSION=$(curl -sX GET "https://api.github.com/repos/coredns/coredns/releases/latest" \
                      | awk '/tag_name/{print $4;exit}' FS='[""]' \
                      | awk '{print substr($1,2); }') && \
    curl -L -o /tmp/coredns.tar.gz "https://github.com/coredns/coredns/releases/download/v${COREDNS_VERSION}/coredns_${COREDNS_VERSION}_linux_amd64.tgz" && \
    tar xf /tmp/coredns.tar.gz -C /

RUN echo "**** Clean up ****" && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*


COPY deploy/corefile.conf /routing/corefile.conf
COPY deploy/coredns.service /etc/systemd/system/coredns.service
COPY deploy/setup.sh /routing/setup.sh

RUN chmod 755 /routing/setup.sh



FROM golang:1.23-bullseye AS build

COPY go.* /src/

WORKDIR /src

RUN go mod download

COPY . .

ARG CMD_PATH=cmd/api
ARG CMD_NAME=main.go
ENV CMD_FULL_PATH=./${CMD_PATH}/${CMD_NAME}

RUN go build -o /app/app $CMD_FULL_PATH



FROM base AS app

ARG CMD_PATH=cmd/api
COPY --from=build /app/. /app
COPY --from=build /src/${CMD_PATH}/config.json /app/

WORKDIR /app

ENTRYPOINT [ "/routing/setup.sh" ]

CMD [ "/app/app" ]

