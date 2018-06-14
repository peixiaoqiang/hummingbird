# Hummingbird
Hummingbird is a big data computing and microservices platform based on Kubernetes. It includes deployment architecture, deployment tools, integration tools for Spark and Kubernetes, and some custom components.

## Getting Started
Run ipallocator server:

```
$ go run pkg/network/allocator/server/ipallocator_server.go -config=etc/ipallocator.conf
```

Grpc with server:

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
## Requirements
### Version
Kubernetes version v1.9.x.
## License
The project is Open Source software released under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0.html).


