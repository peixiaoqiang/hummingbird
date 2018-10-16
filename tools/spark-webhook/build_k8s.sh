#!/usr/bin/env bash
image=$1

function deploy_webhook {
    cat webhook.yaml | sed -e "s|\${IMAGE}|${image}|g" | kubectl apply -f -
}

deploy_webhook