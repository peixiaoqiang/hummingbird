/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package etcd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"path"

	"github.com/TalkingData/hummingbird/pkg/storage"
	etcd "github.com/coreos/etcd/client"
	"github.com/golang/glog"
)

// Creates a new storage interface from the client
// TODO: deprecate in favor of storage.Config abstraction over time
func NewEtcdStorage(client etcd.Client, prefix string, quorum bool) storage.Interface {
	return &etcdHelper{
		etcdMemberAPI: etcd.NewMembersAPI(client),
		etcdKeysAPI:   etcd.NewKeysAPI(client),
		pathPrefix:    path.Join("/", prefix),
		quorum:        quorum,
	}
}

// etcdHelper is the reference implementation of storage.Interface.
type etcdHelper struct {
	etcdMemberAPI etcd.MembersAPI
	etcdKeysAPI   etcd.KeysAPI
	// prefix for all etcd keys
	pathPrefix string
	// if true,  perform quorum read
	quorum bool
}

// Implements storage.Interface.
func (h *etcdHelper) Create(ctx context.Context, key string, obj storage.Object) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}
	key = path.Join(h.pathPrefix, key)
	data, err := encode(obj)
	if err != nil {
		return err
	}
	opts := etcd.SetOptions{
		PrevExist: etcd.PrevNoExist,
	}

	_, err = h.etcdKeysAPI.Set(ctx, key, data, &opts)
	return err
}

func encode(obj storage.Object) (string, error) {
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	err := enc.Encode(obj)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func decode(str string, ptr storage.Object) error {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(ptr)
	return err
}

// Implements storage.Interface.
func (h *etcdHelper) Delete(ctx context.Context, key string) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}

	key = path.Join(h.pathPrefix, key)
	_, err := h.etcdKeysAPI.Delete(ctx, key, nil)
	if err != nil {
		return err
	}
	return nil
}

// Implements storage.Interface.
func (h *etcdHelper) Get(ctx context.Context, key string, objPtr storage.Object) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}

	key = path.Join(h.pathPrefix, key)
	res, err := h.etcdKeysAPI.Get(ctx, key, nil)
	if err != nil {
		return err
	}

	err = decode(res.Node.Value, objPtr)
	if err != nil {
		return err
	}
	return nil
}

// Implements storage.Interface.
func (h *etcdHelper) Update(ctx context.Context, key string, objPtr storage.Object) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}

	data, err := encode(objPtr)
	if err != nil {
		return err
	}

	opts := etcd.SetOptions{
		PrevExist: etcd.PrevExist,
	}
	_, err = h.etcdKeysAPI.Set(ctx, key, data, &opts)
	if err != nil {
		return err
	}
	return nil
}
