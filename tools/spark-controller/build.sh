#!/usr/bin/env bash
src_path=$GOPATH/src/github.com/TalkingData/hummingbird
image=$1

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
    docker push $image
}   

# preinstall
build_image