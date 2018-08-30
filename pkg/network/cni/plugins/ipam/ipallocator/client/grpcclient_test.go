package client

import (
	"sync"
	"testing"

	"github.com/TalkingData/hummingbird/pkg/network/allocator/service"
	"github.com/TalkingData/hummingbird/pkg/utils"
	"github.com/containernetworking/cni/pkg/skel"

	"google.golang.org/grpc"
)

const testServerIP = "172.20.65.3:9000"

func TestAllocateNext(t *testing.T) {
	client, cleanup, err := getTestClient(t)
	if err != nil {
		t.Fatalf("cannot get the client: %v", err)
	}
	defer cleanup()

	ip, err := AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, client)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestRelease(t *testing.T) {
	client, cleanup, err := getTestClient(t)
	if err != nil {
		t.Fatalf("cannot get the client: %v", err)
	}
	defer cleanup()

	testConID := utils.GetRandomString(8)
	ip, err := AllocateNext(&skel.CmdArgs{ContainerID: testConID}, client)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)

	err = Release(&skel.CmdArgs{ContainerID: testConID}, client)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllocateParallel(t *testing.T) {
	client, cleanup, err := getTestClient(t)
	if err != nil {
		t.Fatalf("cannot get the client: %v", err)
	}
	defer cleanup()

	count := 100
	wg := sync.WaitGroup{}
	wg.Add(count)

	do := func() {
		defer wg.Done()
		ip, err := AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, client)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(ip)
	}
	for i := 0; i < count; i++ {
		go do()
	}

	wg.Wait()
}

func getTestClient(t *testing.T) (ipallocatorservice.IPAllocatorClient, func(), error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(testServerIP, opts...)

	if err != nil {
		t.Fatalf("cannot establish the conn: %v", err)
		return nil, nil, err
	}

	return ipallocatorservice.NewIPAllocatorClient(conn), func() {
		if conn != nil {
			err = conn.Close()
			if err != nil {
				t.Fatalf("cannot close the conn: %v", err)
			}
		}
	}, nil
}

func TestAllocateNextMany(t *testing.T) {
	client, cleanup, err := getTestClient(t)
	if err != nil {
		t.Fatalf("cannot get the client: %v", err)
	}
	defer cleanup()

	for i := 0; i < 255; i++ {
		ip, err := AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, client)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(ip)
	}
}
