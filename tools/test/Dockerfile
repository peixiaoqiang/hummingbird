FROM golang:1.10.2
WORKDIR /go/src/github.com/TalkingData/hummingbird/
ARG src_tar
COPY $src_tar .
RUN tar zxvf $src_tar
RUN go build -o ipallocator_server github.com/TalkingData/hummingbird/pkg/network/allocator/server
RUN go build -o ipalloc_testclient github.com/TalkingData/hummingbird/tools/test

FROM centos:7.2.1511
RUN mkdir -p /var/lib/ipallocator/bin/
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/ipallocator_server /usr/local/bin
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/ipalloc_testclient /usr/local/bin
