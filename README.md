# WireguardManager

This microservice application allows you to manage Wireguard VPN (Wireguard network interface and peers connections) through API. 
You can also control speed of traffic for each peer connection. 



## How it works?
The application uses [wgctrl-go](https://github.com/WireGuard/wgctrl-go/) library to access the network interface via a netlink socket. Which allows you to turn on/off the network interface and turn on/off each peer. The traffic control utility (tc) is also used to control speed of traffic.

This microservice is a stateful application. Database stores state of the network interface and each connection. After restarting, saved state is automatically restored.

Also you can enable jwt-based authentication (with a role-based access) to endpoints . 
Works with docker container or on a host. 

Avaliable API interfaces: 
- REST - TODO: Documentation.
- GRPC - will be added later.



## How to use it?

#### App launch stages:
1. The application reads the configuration from a json file and environment variables. 
3. Checking saved state. If it didn't exist, then initialize the new state from the configuration parameters and save it to database.
4. Restore the saved state (turn on Wireguard interface and peers).
5. Launch API server (REST) for interactive management of wireguard interface and peer connections.

#### Configs:
  Before launching application, you can set configuration (or default configuration is applied).
  Сonfiguration can be set either through a config.json file or through environment variables (variables have a higher priority).
  App is looking for a config file in ./configs/config.json. Environment variables are not read from the file, so they need to be set manually (if running on the host), or set via docker attributes (if running in a container).
> Examples of env vars and config file [here](configs/)

> Adding env vars in [my docker-compose file](deploy/docker-compose.yml)

#### Deploy:
- Using docker-compose and image from dockerhub
    ```yml
    version: '3.2'

    services:
      wireguard_manager:
        image: akacyberrat/wireguard_manager

        container_name: wireguard_manager

        ports:
          - "51820:51820/udp"
          - "5000:5000"

        volumes:
          - wireguard_manager_db:/app/db/
          - your/mount/path0:/app/log/
          - your/mount/path1:/app/configs/
          - your/mount/path2:/app/jwt/
          - your/mount/path3:/app/ssl/

        env_file:
          - ../configs/config.env.example

        cap_add:
          - NET_ADMIN
          - SYS_MODULE

    volumes:
      wireguard_manager_db:
    ```

- Using [docker-compose](deploy/docker-compose.yml) and [Dockerfile](deploy/Dockerfile) to build and up. You can use [Makefile](Makefile).

- Using only [Dockerfile](deploy/Dockerfile).

- Build it yourself from the source code (it is not recommended to run on a host without a container, because the application does not take into account existing Wireguard interfaces, and also does not take into account their current state, which can cause errors). You can use [Makefile](Makefile).
  > Before launch yourself build, it is very important to configure the communication of network interfaces. You need to configure traffic forwarding between Wireguard interface and main network interface (usually eth0) so that peers could have access to the external network from your host. I use the CoreDNS utility to [configure forwarding](deploy/).
  Thanks to [Mawthuq-Software](https://github.com/Mawthuq-Software/), the solution and some other from [his project for Wireguard](https://github.com/Mawthuq-Software/Wireguard-Manager-and-API). Also you need to install additional packages on the host, you can see in the [Dockerfile](deploy/Dockerfile).
