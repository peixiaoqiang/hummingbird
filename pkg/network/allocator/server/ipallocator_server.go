package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/TalkingData/hummingbird/pkg/network/allocator"
	"github.com/TalkingData/hummingbird/pkg/network/allocator/ipallocator"
	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
)

var configPath = flag.String("config", "", "The ipallocator server config path")
var config *Config

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
	ipR := &ipallocatorservice.IP{Ip: ipCIDR.String()}
	if config != nil {
		if config.Routes != nil {
			newRs := []*ipallocatorservice.Route{}
			for _, r := range config.Routes {
				newRs = append(newRs, &ipallocatorservice.Route{Dst: r.Dst, Gw: r.Gw})
			}
			ipR.Routes = newRs
		}

		ipR.Gateway = config.Gateway
	}
	return ipR, nil
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
	Routes      []Route  `json:"routes"`
	Gateway     string   `json:"gateway"`
}

type Route struct {
	Dst string `json:"dst"`
	Gw  string `json:"gw"`
}

func initConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	config = &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	err := initConfig(*configPath)
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
