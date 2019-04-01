#!/usr/bin/env bash
dns_server=$1
host_domain=$2
times=$3

for i in `seq $times`
do
dig +short @$dns_server $host_domain | grep -v -e '^$' > /dev/null || echo "fail to resolve "$host_domain
done