#!/usr/bin/env bash
image="spark-controller:$1"
repo_server=$2

go build -o spark-controller $GOPATH/src/github.com/TalkingData/hummingbird/spark/starter.go

docker build -t $image .
docker tag $image $repo_server/$image
docker push $repo_server/$image