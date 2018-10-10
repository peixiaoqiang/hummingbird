#!/usr/bin/env bash
image="spark-controller:$1"
repo_server=$2
src_path=$GOPATH/src/github.com/TalkingData/hummmingbird

function preinstall {
    rm -rf src.tgz
    current_path=$PWD
    cd $src_path
    tar czvf src.tgz .
    mv src.tgz $current_path
    cd $current_path
}

function build_image {
    docker build -t $image .
    docker tag $image $repo_server/$image
    docker push $repo_server/$image
}   