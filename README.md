[![Linux Build Status](https://travis-ci.org/containernetworking/plugins.svg?branch=master)](https://travis-ci.org/containernetworking/plugins)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/kcuubx0chr76ev86/branch/master?svg=true)](https://ci.appveyor.com/project/cni-bot/plugins/branch/master)

# 背景
基于cni的提供的接口开发的ipam插件,主要实现容器的IP地址/网关/DNS/路由存储在etcd库中,集中化管理IP地址,如果使用kubernetes macvlan 插件和该插件配合将实现大二层IP地址统一管理的效果.

### IPAM: IP address allocation(官方)
* `dhcp`: Runs a daemon on the host to make DHCP requests on behalf of the container
* `host-local`: maintains a local database of allocated IPs
* `tdipam`: 负责将分配容器的IP地址存储到etcd中 (该代码库)

### 逻辑导图
![image](https://github.com/TalkingData/hummingbird/blob/master/tdipam.png)


### 我们该如何使用

1. 首先将代码拉取进行go get && go build,如果你不熟悉go的编译，可以现在release二进制版本，将二进制文件拷贝到kubernetes cni插件目录中 例如/opt/cni/bin
2. 将demo目录下的10-macvlan.conf文件拷贝到kuberneres 的cni配置文件目录中 例如/etc/cni/net.d/
3. 首先进行一次IP初始化来确定容器IP的起始和结束范围
`tdipam -init init -start 10.0.0.140 -end 10.0.0.150 -subnet 10.0.0.140/17 -gateway 10.0.0.254 -config /etc/cni/net.d/10-macvlan.conf `


###参考

https://github.com/containernetworking/plugins
https://github.com/containernetworking/cni/blob/master/SPEC.md