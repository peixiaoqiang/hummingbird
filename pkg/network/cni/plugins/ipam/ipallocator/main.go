package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

type IPAMConfig struct {
	ServerIP string `json:"server_ip"`
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

	ip, err := AllocateNext(args, conf.IPAM.ServerIP)
	if err != nil {
		log.Printf("grpc add call failed:%v", err)
		return err
	}

	result := &current.Result{}
	result.CNIVersion = conf.CNIVersion
	ipConfig := &current.IPConfig{}
	_, address, err := net.ParseCIDR(ip.Ip)
	if err != nil {
		log.Printf("incorrect cidr:%v", err)
		return err
	}
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
	return types.PrintResult(result, conf.CNIVersion)
}

// cmdDel is called for DELETE requests
func cmdDel(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return err
	}
	_ = conf

	return Release(args, conf.IPAM.ServerIP)
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports("", "0.1.0", "0.2.0", version.Current()))
}
