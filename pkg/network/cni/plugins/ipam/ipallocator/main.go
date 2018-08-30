package main

import (
	"encoding/json"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	grpcclient "github.com/TalkingData/hummingbird/pkg/network/cni/plugins/ipam/ipallocator/client"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

type IPAMConfig struct {
	ServerIP string         `json:"server_ip"`
	Routes   []*types.Route `json:"routes"`
}

type NetConf struct {
	types.NetConf
	IPAM IPAMConfig `json:"ipam"`
}

func parseConfig(stdin []byte) (*NetConf, error) {
	conf := NetConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		log.Printf("failed to parse network configuration: %v", err)
		return nil, err
	}

	return &conf, nil
}

// cmdAdd is called for ADD requests
func cmdAdd(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return err
	}

	client, cleanup, err := getClient(conf)
	defer cleanup()
	if err != nil {
		log.Printf("cannot get the client: %v", err)
		return err
	}

	ip, err := grpcclient.AllocateNext(args, client)
	if err != nil {
		log.Printf("grpc add call failed:%v", err)
		return err
	}

	result := &current.Result{}
	result.CNIVersion = conf.CNIVersion
	ipConfig := &current.IPConfig{}
	ipR, address, err := net.ParseCIDR(ip.Ip)
	address.IP = ipR
	if err != nil {
		log.Printf("incorrect cidr:%v", err)
		return err
	}
	ipConfig.Version = "4"
	ipConfig.Address = *address
	ipConfig.Gateway = net.ParseIP(ip.Gateway)
	result.IPs = append(result.IPs, ipConfig)

	if ip.Routes != nil {
		for _, r := range ip.Routes {
			_, cidr, err := net.ParseCIDR(r.Dst)
			if err != nil {
				log.Printf("incorrect cidr:%v", err)
				return err
			}
			result.Routes = append(result.Routes, &types.Route{Dst: *cidr, GW: net.ParseIP(r.Gw)})
		}
	}
	if conf.IPAM.Routes != nil {
		result.Routes = append(result.Routes, conf.IPAM.Routes...)
	}

	return types.PrintResult(result, conf.CNIVersion)
}

// cmdDel is called for DELETE requests
func cmdDel(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return err
	}
	_ = conf

	client, cleanup, err := getClient(conf)
	defer cleanup()
	if err != nil {
		log.Printf("cannot get the client: %v", err)
	}

	return grpcclient.Release(args, client)
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports("", "0.1.0", "0.2.0", version.Current()))
}

func getClient(conf *NetConf) (ipallocatorservice.IPAllocatorClient, func(), error) {
	conn, err := newConn(conf.IPAM.ServerIP)
	if err != nil {
		log.Printf("cannot establish the conn: %v", err)
		return nil, nil, err
	}

	return ipallocatorservice.NewIPAllocatorClient(conn), func() {
		if conn != nil {
			err = conn.Close()
			if err != nil {
				log.Printf("cannot close the conn: %v", err)
			}
		}
	}, nil
}

func newConn(serverIp string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	return grpc.Dial(serverIp, opts...)
}
