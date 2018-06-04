# How to use
```
import (
	"testing"
	"net"
)

cidr = "10.0.0.0/24"
cidr, err := net.ParseCIDR(cidr)
r := NewCIDRRange(cidr)

// Allocate a available ip
r.AllocateNext()

// Specify an ip
r.Allocate(net.ParseIP("10.0.0.1"))
```

# How it works
1. Use Bitmap to save ip allocation in memory
2. Snapshot Bitmap and store it to etcd
