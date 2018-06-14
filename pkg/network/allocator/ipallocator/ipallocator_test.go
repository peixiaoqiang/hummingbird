package ipallocator

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/TalkingData/hummingbird/pkg/network/allocator"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
)

func TestAllocate(t *testing.T) {
	cidrStr := "10.0.0.0/16"
	config := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: []string{"http://127.0.0.1:2379"}}
	var ipRangeRegistry allocator.RangeRegistry
	_, ipRange, _ := net.ParseCIDR(cidrStr)
	ipAllocator := NewAllocatorCIDRRange(ipRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewContiguousAllocationMap(max, rangeSpec)
		etcd := allocator.NewEtcd(mem, "/ranges/podips", "/hum/cni/ipregistry", config)
		ipRangeRegistry = etcd
		return etcd
	})
	err := ipRangeRegistry.Init()
	if err != nil {
		t.Fatal(err)
	}

	count := 1000
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			ipAllocator.AllocateNext()
		}()
	}
	wg.Wait()

	fmt.Printf("Used count is %d.", ipAllocator.Used())

	defer func() {
		ipRangeRegistry.Clear()
	}()
}

func BenchmarkAllocate(b *testing.B) {
	cidrStr := "10.0.0.0/16"
	config := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: []string{"http://127.0.0.1:2379"}}
	var ipRangeRegistry allocator.RangeRegistry
	_, ipRange, _ := net.ParseCIDR(cidrStr)
	ipAllocator := NewAllocatorCIDRRange(ipRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewContiguousAllocationMap(max, rangeSpec)
		etcd := allocator.NewEtcd(mem, "/hum/cni/podips", "/hum/cni/ipregistry", config)
		ipRangeRegistry = etcd
		return etcd
	})
	ipRangeRegistry.Init()
	for i := 0; i < b.N; i++ {
		ipAllocator.AllocateNext()
	}

	//fmt.Println(ipAllocator.Used())
}
