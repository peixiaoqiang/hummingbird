package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/TalkingData/hummingbird/pkg/network/allocator"
	"github.com/TalkingData/hummingbird/pkg/network/allocator/ipallocator"
	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/golang/glog"
)

var configPath = flag.String("config", "", "The ipallocator server config path")
var config *Config

type IPAllocatorServer struct {
	IPAllocator   *ipallocator.Range
	IPRegistry    allocator.IPRegistry
	RangeRegistry allocator.RangeRegistry
}

func (s *IPAllocatorServer) AllocateNext(ctx context.Context, ip *ipallocatorservice.IP) (*ipallocatorservice.IP, error) {
	glog.Infof("start to allocate ip for container %s", ip.ContainerID)
	assignedIP, err := s.IPAllocator.AllocateNext()
	glog.V(1).Infof("allocate ip successfully, ip is %v", assignedIP)
	if err != nil {
		glog.Errorf("allocate ip error: %v", err)
		return nil, err
	}

	glog.V(1).Infof("start to register, ip is %v, container_id is %s", assignedIP, ip.ContainerID)
	err = s.IPRegistry.Register(&assignedIP, ip.ContainerID)
	if err != nil {
		glog.Errorf("register ip error: %v", err)
		return nil, err
	}
	glog.V(1).Infof("register ip successfully, ip is %v, container_id is %s", assignedIP, ip.ContainerID)

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
	glog.Infof("allocate and register ip successfully, ip is %v", ipR)
	return ipR, nil
}

func (s *IPAllocatorServer) Release(ctx context.Context, ip *ipallocatorservice.IP) (*ipallocatorservice.Blank, error) {
	glog.Infof("start to release ip, ip is %v", ip)
	ipReg, err := s.IPRegistry.GetIP(ip.ContainerID)
	if err != nil {
		glog.Errorf("fail to get ip %v, error is %v", ip, err)
		return nil, err
	}

	glog.V(1).Infof("start to deregister ip %v", ipReg)
	err = s.IPRegistry.Deregister(ip.ContainerID)
	if err != nil {
		glog.Errorf("fail to deregister ip %v, error is %v", ip, err)
		return nil, err
	}

	glog.V(1).Infof("start release ip %v", ipReg)
	err = s.IPAllocator.Release(*ipReg)
	glog.V(1).Infof("release ip successfully, ip is %v", ipReg)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("release and deregister ip successfully, ip is %v", ip)
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
	defer glog.Flush()

	err := initConfig(*configPath)
	if err != nil {
		glog.Fatalf("fail to load config %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		glog.Fatalf("fail to listen %v", err)
	}
	grpcServer := grpc.NewServer()
	server := &IPAllocatorServer{}
	err = newServer(config, server)
	if err != nil {
		glog.Fatalf("fail to start server %v", err)
	}

	ipallocatorservice.RegisterIPAllocatorServer(grpcServer, server)
	glog.Infof("start server on port %d", config.Port)
	grpcServer.Serve(lis)
}
