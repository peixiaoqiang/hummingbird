FROM golang:1.10.2
WORKDIR $GOPATH/src/github.com/TalkingData/hummingbird/
COPY src.tgz .
RUN tar zxvf src.tgz
RUN go build -o ipallocator_server github.com/TalkingData/hummingbird/pkg/network/allocator/server
RUN go build -o ipallocator github.com/TalkingData/hummingbird/pkg/network/cni/plugins/ipam/ipallocator

FROM centos:7.2.1511
CMD mkdir -p /var/lib/ipallocator/bin/
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/ipallocator /usr/local/ipallocator/bin
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/ipallocator_server /usr/local/bin