# Hummingbird
[![Go Report Card](https://goreportcard.com/badge/github.com/TalkingData/hummingbird)](https://goreportcard.com/report/github.com/TalkingData/hummingbird)
## Overview
Hummingbird is designed to a big data computing and microservices platform based on Kubernetes. It includes deployment architecture, deployment tools, integration tools for Spark and Kubernetes, and some custom components, such as ipallocator, spark controller.
## IPAllocator
IPAllocator is a Kubernetes CNI plugin to manage ips. It uses bitmap mechanism to allocate ip and applies etcd as storeage backend.
### Getting Started
Run ipallocator server:

```
$ go run pkg/network/allocator/server/ipallocator_server.go -config=etc/ipallocator.conf
```

Communicate with server:

```
import (
   "testing"  
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/TalkingData/hummingbird/pkg/utils"
)

const testServerIP = "127.0.0.1:10000"

// Allocate IP
func TestAllocateNext(t *testing.T) {
	ip, err := AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, testServerIP)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

// Release IP
func TestRelease(t *testing.T) {
	testConID := utils.GetRandomString(8)
	ip, err := AllocateNext(&skel.CmdArgs{ContainerID: testConID}, testServerIP)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)

	err = Release(&skel.CmdArgs{ContainerID: testConID}, testServerIP)
	if err != nil {
		t.Fatal(err)
	}
}
```

Install CNI Plugin:

```
$ go build -o /opt/cni/bin/ipallocator github.com/TalkingData/hummingbird/pkg/network/cni/plugins/ipam/ipallocator

$ cat /etc/cni.d/10-macvlan.json
{
  "name": "macvlan",
  "type": "macvlan",
  "master": "eth2",
  "mode": "bridge",
  "ipam": {
    "type": "ipallocator",
    "server_ip": ""
  }
}
```
Or yuo can use yaml file to install it in Kubernetes:

```
$ ./tools/bash.sh <tag> <repo-server>
# Change your yaml
$ kubectl apply -f tools/ipallocator-server.yaml tools/ipallocator-cni.yaml
```
## Spark Controller
Spark Controller is a controller based on Kubernetes Informer that enable to watch spark driver pod to retrieve and store spark application information. There is also a http server to access. You can find more details in [spark-controller](spark/README.md).
## Spark Webhook for Kuberentes
Spark Webhook is a Kubernetes admission webhook service. Admission webhooks are HTTP callbacks that receive admission requests and do something with them. It can dynamically mutate the Spark driver pod. Please find more details in [Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks) and [spark-webhook](spark/webhook/README.md).
## Requirements
### Version
Kubernetes version v1.9.x+.
## License
The project is Open Source software released under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0.html).


