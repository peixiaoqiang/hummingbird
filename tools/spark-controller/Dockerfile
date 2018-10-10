FROM golang:1.10.2
WORKDIR /go/src/github.com/TalkingData/hummingbird
COPY src.tgz .
RUN tar zxvf src.tgz
RUN go build -o spark-controller pkg/spark

FROM centos:7.2.1511
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/spark-controller /usr/local/bin
WORKDIR /usr/local/sparkcontroller
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/html /usr/local/sparkcontroller
