package main

import (
	"github.com/containernetworking/cni/pkg/types"
	"strings"
	"net"
	"log"
)

func GetDns(dnsEtcdConfig *map[string]string) types.DNS {

	dns := types.DNS{}
	for k,v := range *dnsEtcdConfig {
		if strings.Contains(k,"domain"){
			dns.Domain = v
		}

		if strings.Contains(k,"nameservers"){
			ip := net.ParseIP(v)
			if ip != nil {
				dns.Nameservers = append(dns.Nameservers, v)
			}else{
				log.Fatal("Nameserver IP error")
			}
		}

		if strings.Contains(k,"search"){
			dns.Search = append(dns.Search,v)
		}

		if strings.Contains(k,"options"){
			dns.Options= append(dns.Options,v)
		}

	}

	return dns

}
