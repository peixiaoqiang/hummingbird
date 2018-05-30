package main

import (
	//"github.com/containernetworking/cni/pkg/types"
	"fmt"
)

func GetRoute(routeEtcdConfig *map[string]string, RouteRoad *IpamConfig) {
	//var route types.DNS
	//route = &types.DNS{}
	var Rrule []string

	for _, v := range *routeEtcdConfig {
		//Rrule = append(Rrule,v)
		fmt.Println(v)
	}

	fmt.Println(Rrule)

}
