package main

import (
	"sync"
	"testing"

	"github.com/TalkingData/hummingbird/pkg/utils"
	"github.com/containernetworking/cni/pkg/skel"
)

const testServerIP = "127.0.0.1:10000"

func TestAllocateNext(t *testing.T) {
	ip, err := AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, testServerIP)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestRelease(t *testing.T) {
	testConID := utils.GetRandomString(8)
	ip, err := AllocateNext(&skel.CmdArgs{ContainerID: testConID}, testServerIP)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)

	err = Release(&skel.CmdArgs{ContainerID: testConID}, testServerIP)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllocateParallel(t *testing.T) {
	count := 1000
	wg := sync.WaitGroup{}
	wg.Add(count)

	do := func() {
		defer wg.Done()
		ip, err := AllocateNext(&skel.CmdArgs{ContainerID: utils.GetRandomString(8)}, testServerIP)
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
