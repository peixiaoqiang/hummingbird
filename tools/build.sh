#!/usr/bin/env bash
image="ipallocator:$1"
repo_server=$2

go build -o ipallocator_server github.com/TalkingData/hummingbird/pkg/network/allocator/server
go build -o ipallocator github.com/TalkingData/hummingbird/pkg/network/cni/plugins/ipam/ipallocator

docker build -t $image .
docker tag $image $repo_server/$image
docker push $repo_server/$image