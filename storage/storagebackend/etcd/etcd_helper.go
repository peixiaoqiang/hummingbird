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
	"context"
	"path"
	etcd "github.com/coreos/etcd/client"
	"github.com/TalkingData/hummingbird/storage"
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
func (h *etcdHelper) Create(ctx context.Context, key string, obj, out storage.Object, ttl uint64) error {
	// TODO
	return nil
}

// Implements storage.Interface.
func (h *etcdHelper) Delete(ctx context.Context, key string, out storage.Object) error {
	// TODO
	return nil
}

// Implements storage.Interface.
func (h *etcdHelper) Get(ctx context.Context, key string, objPtr storage.Object) error {
	// TODO
	return nil
}
