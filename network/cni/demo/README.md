[![Linux Build Status](https://travis-ci.org/containernetworking/plugins.svg?branch=master)](https://travis-ci.org/containernetworking/plugins)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/kcuubx0chr76ev86/branch/master?svg=true)](https://ci.appveyor.com/project/cni-bot/plugins/branch/master)

# 10-macvlan.conf 配置文件解释
`{
        "name": "macvlannet",   
        "type": "macvlan",
        "master": "ens33",
        "mode": "bridge",
        "isGateway": true,
        "ipMasq": false,
        "ipam": {
                "type": "tdipam",  //选择tdipam插件
                "etcdcluster": "http://127.0.0.1:4379,http://10.0.0.65:2379", //etcd服务器地址
                "nodenetwork": "/prod/nodenetwork/",  //在etcd中宿主机网络范围key路劲
                "alreadyusedip": "/prod/alreadyusedip/",  //在etcd中容器网络范围key路劲
                "containernetwork": "/prod/containernetwork/", //在etcd中容器网络范围key路劲
                "routes": "/prod/routes/", //在etcd中路由key路劲 
                "dns": "/prod/dns/" //在etcd中dns服务器key路劲 
        }
}`


