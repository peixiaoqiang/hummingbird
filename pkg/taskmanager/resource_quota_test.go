package taskmanager

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"

	"github.com/TalkingData/hummingbird/pkg/storage"
)

type FakeStorage struct {
	s map[string]string
}

func (s *FakeStorage) Create(ctx context.Context, key string, obj storage.Object) error {
	v, _ := s.encode(obj)
	s.s[key] = v
	return nil
}
func (s *FakeStorage) Update(ctx context.Context, key string, obj storage.Object) error {
	s.Create(ctx, key, obj)
	return nil
}
func (s *FakeStorage) CreateOrUpdate(ctx context.Context, key string, obj storage.Object) error {
	s.Create(ctx, key, obj)
	return nil
}
func (s *FakeStorage) Delete(ctx context.Context, key string) error {
	delete(s.s, key)
	return nil
}
func (s *FakeStorage) Get(ctx context.Context, key string, objPtr storage.Object) error {
	v := s.s[key]
	s.decode(v, objPtr)
	return nil
}
func (s *FakeStorage) Lock(ctx context.Context) error {
	return nil
}
func (s *FakeStorage) Unlock(ctx context.Context) error {
	return nil
}

func (e *FakeStorage) encode(obj storage.Object) (string, error) {
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	err := enc.Encode(obj)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func (e *FakeStorage) decode(str string, ptr storage.Object) error {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(ptr)
	return err
}

func NewFakeResourceQuotaManager(conf *Conf) (*ResourceQuotaManager, error) {
	return &ResourceQuotaManager{storage: &FakeStorage{s: map[string]string{}}, conf: conf, context: context.TODO()}, nil
}

// import "testing"

// func testResQuotaConf() *Conf {
// 	conf := clone(CONF)
// 	conf.StorageKeyPrefix = "testtaskmanager"
// 	conf.EtcdIps = []string{"localhost:2379"}
// 	conf.ResourceQuotaName = "test-spark"
// 	return conf
// }

// func TestInit(t *testing.T) {
// 	conf := testResQuotaConf()
// 	m, err := NewResourceQuotaManager(conf)
// 	if err != nil {
// 		t.Fatalf("%v", err)
// 	}
// 	quota := ResourceQuota{LimitCPU: 1, LimitMem: 100}
// 	err = m.Init(conf.ResourceQuotaName, &quota)
// 	if err != nil {
// 		t.Fatalf("%v", err)
// 	}
// }
