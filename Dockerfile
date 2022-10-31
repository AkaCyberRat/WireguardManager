FROM wg_control_base
WORKDIR /usr/src/wireguard_manager
COPY . .

RUN \
 mv additional/Corefile /app/Corefile &&\
 mv additional/coredns.service /etc/systemd/system/ &&\
 systemctl daemon-reload &&\
 chmod 777 additional/SetupIptableAndRunApp.sh &&\
 go build main.go 

EXPOSE 51830/udp
CMD [ "./additional/SetupIptableAndRunApp.sh" ]
#ENTRYPOINT ["./main"]