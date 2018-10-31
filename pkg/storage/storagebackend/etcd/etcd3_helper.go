package etcd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"path"

	"github.com/TalkingData/hummingbird/pkg/storage"
	v3 "go.etcd.io/etcd/clientv3"
)

func NewEtcd3Storage(client *v3.Client, prefix string) storage.Interface {
	return &etcd3Helper{client, prefix}
}

type etcd3Helper struct {
	client *v3.Client
	prefix string
}

func (e *etcd3Helper) encode(obj storage.Object) (string, error) {
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	err := enc.Encode(obj)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func (e *etcd3Helper) decode(str string, ptr storage.Object) error {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(ptr)
	return err
}

func (e *etcd3Helper) getKey(key string) string {
	return path.Join(e.prefix, key)
}

func (e *etcd3Helper) Create(ctx context.Context, key string, obj storage.Object) error {
	val, err := e.encode(obj)
	if err != nil {
		return err
	}
	cmp := v3.Compare(v3.Version(e.getKey(key)), "=", 0)
	req := v3.OpPut(e.getKey(key), val, v3.WithLease(v3.NoLease))
	txnresp, err := e.client.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !txnresp.Succeeded {
		return storage.ErrKeyExists
	}
	return nil
}

func (e *etcd3Helper) Update(ctx context.Context, key string, obj storage.Object) error {
	val, err := e.encode(obj)
	if err != nil {
		return err
	}
	cmp := v3.Compare(v3.Version(key), ">", 0)
	req := v3.OpPut(e.getKey(key), val, v3.WithLease(v3.NoLease))
	txnresp, err := e.client.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !txnresp.Succeeded {
		return storage.ErrKeyNoExists
	}
	return nil
}

func (e *etcd3Helper) CreateOrUpdate(ctx context.Context, key string, obj storage.Object) error {
	val, err := e.encode(obj)
	if err != nil {
		return err
	}
	_, err = e.client.Put(ctx, e.getKey(key), val, v3.WithLease(v3.NoLease))
	return err
}

func (e *etcd3Helper) Delete(ctx context.Context, key string) error {
	_, err := e.client.Delete(context.TODO(), e.getKey(key))
	return err
}

func (e *etcd3Helper) Get(ctx context.Context, key string, objPtr storage.Object) error {
	resp, err := e.client.Get(ctx, e.getKey(key), v3.WithLease(v3.NoLease))
	if err != nil {
		return err
	}
	if resp.Count > 0 {
		val := resp.Kvs[0].Value
		err := e.decode(string(val), objPtr)
		if err != nil {
			return err
		}
		return nil
	}
	return storage.ErrKeyNoExists
}

func (e *etcd3Helper) Lock(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (e *etcd3Helper) Unlock(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
