package main

import (
	"sync"
	"flag"

	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	grpcclient "github.com/TalkingData/hummingbird/pkg/network/cni/plugins/ipam/ipallocator/client"
	"github.com/TalkingData/hummingbird/pkg/utils"
	"github.com/containernetworking/cni/pkg/skel"

	"google.golang.org/grpc"
	"github.com/golang/glog"
	"time"
)

var (
	serverIP    = flag.String("server_ip", "localhost:9000", "The ipallocator ip")
	method      = flag.String("method", "AllocateNext", "The method for test")
	concurrentN = flag.Int("concurrent_num", 100, "Concurrent num for test worker")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	client, cleanup, err := getTestClient()

	if err != nil {
		glog.Fatalf("cannot get the client: %v", err)
	}
	defer cleanup()

	wg := sync.WaitGroup{}
	wg.Add(*concurrentN)

	start := time.Now()
	for i := 0; i < *concurrentN; i++ {
		go func() {
			defer wg.Done()
			if err := call(client); err != nil {
				glog.Errorf("fail to call %v, error is %v", method, err)
			}
		}()
	}

	wg.Wait()
	end := time.Now()
	glog.Infof("finish test, cost time %v", end.Unix()-start.Unix())
}

func call(client ipallocatorservice.IPAllocatorClient) (err error) {
	switch *method {
	case "AllocateNext":
		ip, err := grpcclient.AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, client)
		if err == nil {
			glog.Infoln(ip.Ip)
		}
	default:
	}

	if err != nil {
		return err
	}
	return nil
}

func getTestClient() (ipallocatorservice.IPAllocatorClient, func(), error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverIP, opts...)

	if err != nil {
		glog.Fatalf("cannot establish the conn: %v", err)
		return nil, nil, err
	}

	return ipallocatorservice.NewIPAllocatorClient(conn), func() {
		if conn != nil {
			err = conn.Close()
			if err != nil {
				glog.Fatalf("cannot close the conn: %v", err)
			}
		}
	}, nil
}
