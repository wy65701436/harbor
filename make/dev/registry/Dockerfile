FROM golang:1.9.2

MAINTAINER wangyan@vmware.com

COPY . /go/src/github.com/vmware/harbor

WORKDIR /go/src/github.com/vmware/harbor/src/registry

RUN go build -a -o /go/bin/harbor_registry \
    && chmod u+x /go/bin/harbor_registry 
WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor_registry"]
