#!/usr/bin/env bash
image=$1
current_path="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

function preinstall {
    rm -rf src.tgz
    cd $root_path
    tar czvf $current_path/src.tgz . --exclude=.git
    cd $current_path
}

function build_image {
    docker build -t $image --build-arg work_dir=$current_path .
    docker push $image
}

source ../../env
preinstall
build_image