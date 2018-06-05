[![Linux Build Status](https://travis-ci.org/containernetworking/plugins.svg?branch=master)](https://travis-ci.org/containernetworking/plugins)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/kcuubx0chr76ev86/branch/master?svg=true)](https://ci.appveyor.com/project/cni-bot/plugins/branch/master)

# 背景
基于cni的提供的接口开发的ipam插件,主要实现容器的IP地址/网关/DNS/路由存储在etcd库中,集中化管理IP地址,如果使用kubernetes macvlan 插件和该插件配合将实现大二层IP地址统一管理的效果.

## Plugins supplied:(官方)
### Main: interface-creating
* `bridge`: Creates a bridge, adds the host and the container to it.
* `ipvlan`: Adds an [ipvlan](https://www.kernel.org/doc/Documentation/networking/ipvlan.txt) interface in the container
* `loopback`: Creates a loopback interface
* `macvlan`: Creates a new MAC address, forwards all traffic to that to the container
* `ptp`: Creates a veth pair.
* `vlan`: Allocates a vlan device.


### IPAM: IP address allocation(官方)
* `dhcp`: Runs a daemon on the host to make DHCP requests on behalf of the container
* `host-local`: maintains a local database of allocated IPs
* `tdipam`: 负责将分配容器的IP地址存储到etcd中 (该代码库)

### Meta: other plugins
* `flannel`: generates an interface corresponding to a flannel config file
* `tuning`: Tweaks sysctl parameters of an existing interface
* `portmap`: An iptables-based portmapping plugin. Maps ports from the host's address space to the container.

### 逻辑导图
![image](https://github.com/TalkingData/hummingbird/blob/master/network/cni/tdipam.png)


### 我们该如何使用

1.  cd plugins && ./build.sh 将会在bin目录下产生tdipam,将编译出来的二进制文件拷贝到kubernetes cni插件目录中 例如/opt/cni/bin
2. 将demo目录下的10-macvlan.conf文件拷贝到kuberneres 的cni配置文件目录中 例如/etc/cni/net.d/
3. 首先进行一次IP初始化来确定容器IP的起始和结束范围
tdipam -init init -start 10.0.0.140 -end 10.0.0.150 -subnet 10.0.0.140/17 -gateway 10.0.0.254 -nameservers 8.8.8.8,4.4.4.4 -defaultroute 172.20.0.1,0.0.0.0/0 -config /etc/cni/net.d/10-macvlan.conf
4. 下面我们可以正常用啦
