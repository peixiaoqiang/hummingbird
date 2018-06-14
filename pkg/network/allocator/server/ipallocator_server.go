package main

import (
	"net"
	"flag"
	"fmt"
	"log"
	"io/ioutil"
	"encoding/json"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/network/allocator/ipallocator"
	"github.com/TalkingData/hummingbird/pkg/network/allocator"
	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
)

var (
	configPath = flag.String("config", "", "The ipallocator server config path")
)

type IPAllocatorServer struct {
	IPAllocator   *ipallocator.Range
	IPRegistry    allocator.IPRegistry
	RangeRegistry allocator.RangeRegistry
}

func (s *IPAllocatorServer) AllocateNext(ctx context.Context, ip *ipallocatorservice.IP) (*ipallocatorservice.IP, error) {
	assignedIP, err := s.IPAllocator.AllocateNext()
	if err != nil {
		return nil, err
	}

	err = s.IPRegistry.Register(&assignedIP, ip.ContainerID)
	if err != nil {
		return nil, err
	}

	ipCIDR := net.IPNet{IP: assignedIP, Mask: s.IPAllocator.CIDR().Mask}
	return &ipallocatorservice.IP{Ip: ipCIDR.String()}, nil
}

func (s *IPAllocatorServer) Release(ctx context.Context, ip *ipallocatorservice.IP) (*ipallocatorservice.Blank, error) {
	ipReg, err := s.IPRegistry.GetIP(ip.ContainerID)
	if err != nil {
		return nil, err
	}

	err = s.IPRegistry.Deregister(ip.ContainerID)
	if err != nil {
		return nil, err
	}

	err = s.IPAllocator.Release(*ipReg)
	if err != nil {
		return nil, err
	}

	return &ipallocatorservice.Blank{}, nil
}

func newServer(config *Config, server *IPAllocatorServer) error {
	storeConfig := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: config.EtcdIps}
	var rangeRegistry allocator.RangeRegistry
	var ipRegistry allocator.IPRegistry
	_, ipRange, _ := net.ParseCIDR(config.CIDR)
	ipAllocator := ipallocator.NewAllocatorCIDRRange(ipRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewContiguousAllocationMap(max, rangeSpec)
		etcd := allocator.NewEtcd(mem, config.BaseKey, config.RegistryKey, storeConfig)
		rangeRegistry = etcd
		ipRegistry = etcd
		return etcd
	})

	err := rangeRegistry.Init()
	if err != nil {
		return err
	}

	server.IPAllocator = ipAllocator
	server.RangeRegistry = rangeRegistry
	server.IPRegistry = ipRegistry
	return nil
}

type Config struct {
	Port        int      `json:"port"`
	CIDR        string   `json:"cidr"`
	BaseKey     string   `json:"base_key"`
	RegistryKey string   `json:"registry_key"`
	EtcdIps     []string `json:"etcd_ips"`
}

func loadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	flag.Parse()
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failted to load config:%v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	server := &IPAllocatorServer{}
	err = newServer(config, server)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	ipallocatorservice.RegisterIPAllocatorServer(grpcServer, server)
	grpcServer.Serve(lis)
}
