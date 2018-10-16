FROM golang:1.10.2
WORKDIR /go/src/github.com/TalkingData/hummingbird
COPY src.tgz .
RUN tar zxvf src.tgz
RUN go build -o spark-webhook github.com/TalkingData/hummingbird/spark/webhook

FROM centos:7.2.1511
COPY --from=0 /go/src/github.com/TalkingData/hummingbird/spark-webhook /usr/local/bin