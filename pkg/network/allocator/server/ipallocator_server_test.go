package main

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	"github.com/TalkingData/hummingbird/pkg/utils"
)

func getTestServer(t *testing.T, server *IPAllocatorServer) func() {
	keyPrefix := "/" + utils.GetRandomString(8)
	config := &Config{CIDR: "10.0.0.0/24", Port: 1000, BaseKey: keyPrefix + "/podips", RegistryKey: keyPrefix + "/ipregistry", EtcdIps: []string{"http://127.0.0.1:2379"}}
	err := newServer(config, server)
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		server.IPRegistry.ClearIPRegistry()
		server.RangeRegistry.ClearRangeRegistry()
	}
}

func TestAllocateNext(t *testing.T) {
	server := &IPAllocatorServer{}
	defer getTestServer(t, server)()

	ip := &ipallocatorservice.IP{ContainerID: utils.GetRandomString(8)}
	ip, err := server.AllocateNext(context.TODO(), ip)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ip)
}

func TestRelease(t *testing.T) {
	server := &IPAllocatorServer{}
	defer getTestServer(t, server)()

	ip := &ipallocatorservice.IP{ContainerID: utils.GetRandomString(8)}
	server.AllocateNext(context.TODO(), ip)
	_, err := server.Release(context.TODO(), ip)
	if err != nil {
		t.Fatal(err)
	}
}
