package queue

import (
	"testing"

	v3 "go.etcd.io/etcd/clientv3"
)

func TestPutNewKey(t *testing.T) {
	conf := v3.Config{Endpoints: []string{"localhost:2379"}}
	client, err := v3.New(conf)
	_, err = putNewKV(client, "testtaskmanager/resourcequota/test-spark", "test", v3.NoLease)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
