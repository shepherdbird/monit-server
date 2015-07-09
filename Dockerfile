FROM golang:1.3
MAINTAINER Sun Jianbo <wonderflow.sun@gmail.com> (@wonderflow)


ENV GOPATH /go:/go/src/github.com/shepherdbird/monit-server/Godeps/_workspace
COPY . /go/src/github.com/shepherdbird/monit-server
WORKDIR /go/src/github.com/shepherdbird/monit-server
ENTRYPOINT ["build/build"]