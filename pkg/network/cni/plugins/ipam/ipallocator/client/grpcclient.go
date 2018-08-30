package client

import (
	"time"

	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	"github.com/containernetworking/cni/pkg/skel"
	"golang.org/x/net/context"
	"github.com/golang/glog"
)

func AllocateNext(args *skel.CmdArgs, client ipallocatorservice.IPAllocatorClient) (*ipallocatorservice.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ip := &ipallocatorservice.IP{ContainerID: args.ContainerID}
	ip, err := client.AllocateNext(ctx, ip)
	if err != nil {
		glog.Errorf("grpc call failed: %v", err)
		return nil, err
	}

	return ip, nil
}

func Release(args *skel.CmdArgs, client ipallocatorservice.IPAllocatorClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := &ipallocatorservice.IP{ContainerID: args.ContainerID}
	client.Release(ctx, ip)
	return nil
}
