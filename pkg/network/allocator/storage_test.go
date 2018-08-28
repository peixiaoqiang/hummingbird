package allocator

import (
	"testing"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
)

func baseTest(t testing.TB, do func(storage Interface)) {
	cidr := "10.0.0.0/16"
	config := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: []string{"http://127.0.0.1:2379"}}
	mem := NewContiguousAllocationMap(10000, cidr)
	etcd := NewEtcd(mem, "/ranges/podips", "/hum/cni/ipregistry", config)
	err := etcd.Init()
	if err != nil {
		t.Fatal(err)
	}
	do(etcd)
	etcd.ClearIPRegistry()
	etcd.ClearRangeRegistry()
}

func TestAllocateNext(t *testing.T) {
	baseTest(t, func(storage Interface) {
		_, _, err := storage.AllocateNext()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func BenchmarkAllocateNext(b *testing.B) {
	baseTest(b, func(storage Interface) {
		for i := 0; i < b.N; i++ {
			_, _, err := storage.AllocateNext()
			if err != nil {
				b.Error(err)
			}
		}
	})

}
