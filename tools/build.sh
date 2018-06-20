#!/usr/bin/env bash
tag="ipallocator:$1"
go build -o ipallocator_server github.com/TalkingData/hummingbird/pkg/network/allocator/server
go build -o ipallocator github.com/TalkingData/hummingbird/pkg/network/cni/plugins/ipam/ipallocator
docker build -t $tag .
docker tag $tag repo_server/$tag
docker push repo_server/$tag