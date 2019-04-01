#!/usr/bin/env bash
image=$1

function preinstall {
    rm -rf src.tgz
    current_path=$PWD
    cd $root_path
    tar czvf src.tgz -C $current_path
    cd $current_path
}

function build_image {
    docker build -t $image .
    docker push $image
}   

source ../env
preinstall
build_image