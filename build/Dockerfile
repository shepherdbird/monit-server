FROM progrium/busybox
MAINTAINER Sun Jianbo <wonderflow.sun@gmail.com> (@wonderflow)

# Grab monit-server after build successfully.
ADD monit-server /usr/bin/monit-server
ADD cert.pem /usr/bin/cert.pem
ADD key.pem /usr/bin/key.pem
ADD iptables /usr/bin/iptables
ADD curl /usr/bin/curl
WORKDIR /usr/bin
ENTRYPOINT ["./monit-server"]