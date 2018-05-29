package main

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ip"
	"log"
	"net"
)

type Range struct {
	RangeStart net.IP
	RangeEnd   net.IP
	Subnet     types.IPNet
	Gateway    net.IP
	ip         []byte
}

// 给定一个IP地址的范围
func (r *Range) Canonicalize() error {
	if err := canonicalizeIP(&r.Subnet.IP); err != nil {
		return err
	}

	//不能创建没有网络地址的范围例如掩码为/31和32 位
	ones, masklen := r.Subnet.Mask.Size()

	if ones > masklen-2 {
		return fmt.Errorf("Network %s too small to allocate from", (*net.IPNet)(&r.Subnet).String())
	}

	if len(r.Subnet.IP) != len(r.Subnet.Mask) {
		return fmt.Errorf("IPNet IP and Mask version mismatch")
	}

	// If the gateway is nil, claim .1
	if r.Gateway == nil {
		r.Gateway = ip.NextIP(r.Subnet.IP)
	} else {
		if err := canonicalizeIP(&r.Gateway); err != nil {
			return err
		}
		subnet := (net.IPNet)(r.Subnet)
		if !subnet.Contains(r.Gateway) {
			return fmt.Errorf("gateway %s not in network %s", r.Gateway.String(), subnet.String())
		}
	}

	// RangeStart: If specified, make sure it's sane (inside the subnet),
	// otherwise use the first free IP (i.e. .1) - this will conflict with the
	// gateway but we skip it in the iterator
	if r.RangeStart != nil {
		if err := canonicalizeIP(&r.RangeStart); err != nil {
			return err
		}

		if !r.Contains(r.RangeStart) {
			return fmt.Errorf("RangeStart %s not in network %s", r.RangeStart.String(), (*net.IPNet)(&r.Subnet).String())
		}
	} else {
		r.RangeStart = ip.NextIP(r.Subnet.IP)
	}

	// RangeEnd: If specified, verify sanity. Otherwise, add a sensible default
	// (e.g. for a /24: .254 if IPv4, ::255 if IPv6)
	if r.RangeEnd != nil {
		if err := canonicalizeIP(&r.RangeEnd); err != nil {
			return err
		}

		if !r.Contains(r.RangeEnd) {
			return fmt.Errorf("RangeEnd %s not in network %s", r.RangeEnd.String(), (*net.IPNet)(&r.Subnet).String())
		}
	} else {
		r.RangeEnd = lastIP(r.Subnet)
	}

	return nil
}

// IsValidIP checks if a given ip is a valid, allocatable address in a given Range
func (r *Range) Contains(addr net.IP) bool {
	if err := canonicalizeIP(&addr); err != nil {
		return false
	}

	subnet := (net.IPNet)(r.Subnet)

	// Not the same address family
	if len(addr) != len(r.Subnet.IP) {
		return false
	}

	// Not in network
	if !subnet.Contains(addr) {
		return false
	}

	// We ignore nils here so we can use this function as we initialize the range.
	if r.RangeStart != nil {
		// Before the range start
		if ip.Cmp(addr, r.RangeStart) < 0 {
			return false
		}
	}

	if r.RangeEnd != nil {
		if ip.Cmp(addr, r.RangeEnd) > 0 {
			// After the  range end
			return false
		}
	}

	return true
}

func (r *Range) String() string {
	return fmt.Sprintf("%s-%s", r.RangeStart.String(), r.RangeEnd.String())
}

// canonicalizeIP makes sure a provided ip is in standard form
func canonicalizeIP(ip *net.IP) error {
	if ip.To4() != nil {
		*ip = ip.To4()
		return nil
	} else if ip.To16() != nil {
		*ip = ip.To16()
		return nil
	}
	return fmt.Errorf("IP %s not v4 nor v6", *ip)
}

func (r *Range) Overlaps(r1 *Range) bool {
	// different familes
	if len(r.RangeStart) != len(r1.RangeStart) {
		return false
	}

	return r.Contains(r1.RangeStart) ||
		r.Contains(r1.RangeEnd) ||
		r1.Contains(r.RangeStart) ||
		r1.Contains(r.RangeEnd)
}

// Determine the last IP of a subnet, excluding the broadcast if IPv4
func lastIP(subnet types.IPNet) net.IP {
	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}
	if subnet.IP.To4() != nil {
		end[3]--
	}

	return end
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func (R *Range) Container(ContainerR *map[string]string, Key string) {
	RangeStart := net.ParseIP((*ContainerR)[Key+"RangeStart"])
	err := canonicalizeIP(&RangeStart)
	if err != nil {
		log.Fatal("Incorrect rangStart address")
	}
	RangeEnd := net.ParseIP((*ContainerR)[Key+"RangeEnd"])
	err = canonicalizeIP(&RangeEnd)
	if err != nil {
		log.Fatal("Incorrect rangEnd address")
	}
	ip, Subnet, err := net.ParseCIDR((*ContainerR)[Key+"SubNet"])
	if err != nil {
		log.Fatal("Incorrect rangEnd SubNet")
	}
	Gateway := net.ParseIP((*ContainerR)[Key+"GateWay"])
	err = canonicalizeIP(&Gateway)
	if err != nil {
		log.Fatal("Incorrect gateway address")
	}
	R.RangeStart = RangeStart
	R.RangeEnd = RangeEnd
	R.ip = ip
	R.Subnet.IP = Subnet.IP
	R.Subnet.Mask = Subnet.Mask
	R.Gateway = Gateway

}

func (R *Range) RangeSet(AlreadUsedIp *map[string]string, IpList *[]string, Key string) (ip []net.IP) {
	var AvailableIpList []net.IP
	for _, Ip := range *IpList {
		if _, ok := (*AlreadUsedIp)[Key+Ip]; !ok {
			ip := net.ParseIP(Ip)
			if R.Contains(ip) == true {
				AvailableIpList=append(AvailableIpList,ip)
			}

		}

	}

	return AvailableIpList
}

func ContainerSearch(AlreadUsedIp *map[string]string, ContainerID string) string {
	for k, v := range *AlreadUsedIp {
		if v == ContainerID {
			return k
		}
	}

	return ""
}
