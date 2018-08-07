#!/usr/bin/env bash
image="spark-watcher:$1"
repo_server=$2

go build -o spark-watcher github.com/TalkingData/hummingbird/pkg/server/starter.go

docker build -t $image .
docker tag $image $repo_server/$image
docker push $repo_server/$image