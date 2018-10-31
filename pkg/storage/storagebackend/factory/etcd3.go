package factory

import (
	"github.com/TalkingData/hummingbird/pkg/storage"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/etcd"
	v3 "go.etcd.io/etcd/clientv3"
)

func newETCD3Storage(c storagebackend.Config) (storage.Interface, DestroyFunc, error) {
	conf := v3.Config{Endpoints: c.ServerList}
	client, err := v3.New(conf)
	if err != nil {
		return nil, nil, err
	}
	return etcd.NewEtcd3Storage(client, c.Prefix), nil, nil
}
