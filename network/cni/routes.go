package main

import (

	"github.com/containernetworking/cni/pkg/types"
	"strings"
	"net"
	"log"
)

func GetRoute(routeEtcdConfig *map[string]string) []*types.Route {
	Routes := []*types.Route{}

	for k,v := range *routeEtcdConfig{
		Route := &types.Route{}
		if strings.Contains(k,"routes"){
			_, dstmask, err := net.ParseCIDR(v)
			if err != nil{
				log.Fatal("Routing destination address error")
			}
			Route.Dst.IP = dstmask.IP
			Route.Dst.Mask = dstmask.Mask
			sk := strings.Split(k,"/")
			k = sk[len(sk)-1]
			Route.GW = net.ParseIP(k)
			Routes = append(Routes,Route)

			return Routes
		}

	}
	return Routes

}
