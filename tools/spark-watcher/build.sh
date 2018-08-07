#!/usr/bin/env bash
image="spark-watcher:$1"
repo_server=$2

go build -o spark-watcher $GOPATH/src/github.com/TalkingData/hummingbird/server/starter.go

docker build -t $image .
docker tag $image $repo_server/$image
docker push $repo_server/$image