package main

import (
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	"github.com/containernetworking/cni/pkg/skel"
)

func newConn(serverIp string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	return grpc.Dial(serverIp, opts...)
}

func getClient(serverIp string) (ipallocatorservice.IPAllocatorClient, func(), error) {
	conn, err := newConn(serverIp)
	if err != nil {
		log.Printf("cannot establish the conn: %v", err)
		return nil, nil, err
	}

	return ipallocatorservice.NewIPAllocatorClient(conn), func() {
		conn.Close()
	}, nil
}

func AllocateNext(args *skel.CmdArgs, serverIp string) (*ipallocatorservice.IP, error) {
	client, cleanup, err := getClient(serverIp)
	defer cleanup()

	if err != nil {
		log.Printf("cannot get client: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := &ipallocatorservice.IP{ContainerID: args.ContainerID}
	ip, err = client.AllocateNext(ctx, ip)
	if err != nil {
		log.Printf("grpc call failed: %v", err)
		return nil, err
	}

	return ip, nil
}

func Release(args *skel.CmdArgs, serverIp string) error {
	client, cleanup, err := getClient(serverIp)
	defer cleanup()

	if err != nil {
		log.Printf("cannot get client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := &ipallocatorservice.IP{ContainerID: args.ContainerID}
	_, err = client.Release(ctx, ip)
	return err
}
