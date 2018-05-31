package main

import (
	"github.com/containernetworking/cni/pkg/types"
	"strings"
	)

func GetDns(dnsEtcdConfig *map[string]string) types.DNS {

	dns := types.DNS{}
	for k,v := range *dnsEtcdConfig {
		if strings.Contains(k,"domain"){
			dns.Domain = v
		}

		if strings.Contains(k,"nameservers"){
			dns.Nameservers = append(dns.Nameservers,v)
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
