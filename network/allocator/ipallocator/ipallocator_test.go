package ipallocator

import (
	"testing"
	"net"
	"github.com/TalkingData/hummingbird/network/allocator"
	"github.com/TalkingData/hummingbird/storage/storagebackend"
	"fmt"
	"encoding/gob"
	"bytes"
)

func TestAllocate(t *testing.T) {
	cidrStr := "10.0.0.0/24"
	config := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: []string{"http://127.0.0.1:2379"}}
	var ipRangeRegistry allocator.RangeRegistry
	_, ipRange, _ := net.ParseCIDR(cidrStr)
	ipAllocator := NewAllocatorCIDRRange(ipRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewContiguousAllocationMap(max, rangeSpec)
		etcd := allocator.NewEtcd(mem, "/ranges/podips", config)
		ipRangeRegistry = etcd
		return etcd
	})
	err := ipRangeRegistry.Init()
	if err != nil {
		t.Fatal(err)
	}

	nextIP, err := ipAllocator.AllocateNext()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(nextIP)
}

func TestEncode(t *testing.T) {
	r := allocator.RangeAllocation{Range: "10.0.0.0/24", Data: []byte{1}}
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	err := enc.Encode(r)
	fmt.Println(out.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	newR := &allocator.RangeAllocation{}
	dec := gob.NewDecoder(bytes.NewBufferString(out.String()))
	err = dec.Decode(newR)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(newR)
}
