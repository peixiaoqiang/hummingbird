package main

import (
	"github.com/containernetworking/cni/pkg/types"
	"strings"
)

func GetDns(dnsEtcdConfig *map[string]string, DnsRoad *IpamConfig) types.DNS {

	var dns types.DNS
	dns = types.DNS{}

	if _, ok := (*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"domain"]; ok {
		dns.Domain = (*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"domain"]

	}
	if _, ok := (*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"nameservers"]; ok {
		var nameservers []string = strings.Split((*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"nameservers"], ",")
		dns.Nameservers = nameservers

	}

	if _, ok := (*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"search"]; ok {
		var search []string = strings.Split((*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"search"], ",")
		dns.Search = search
	}

	if _, ok := (*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"options"]; ok {
		var options []string = strings.Split((*dnsEtcdConfig)[DnsRoad.Ipam.Dns+"options"], ",")
		dns.Options = options
	}
	return dns

}
