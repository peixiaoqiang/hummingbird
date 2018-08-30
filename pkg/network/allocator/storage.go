/*
Copyright 2015 The Kubernetes Authors.

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

package allocator

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/TalkingData/hummingbird/pkg/storage"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/factory"
)

var (
	errorUnableToAllocate = errors.New("unable to allocate")
)

// Etcd exposes a service.Allocator
// TODO: allow multiple allocations to be tried at once
// TODO: subdivide the keyspace to reduce conflicts
// TODO: investigate issuing a CAS without reading first
type Etcd struct {
	lock sync.Mutex

	alloc   Snapshottable
	storage storage.Interface
	last    string

	registryKey string
	baseKey     string
}

// Etcd implements allocator.Interface and rangeallocation.RangeRegistry
var _ Interface = &Etcd{}
var _ RangeRegistry = &Etcd{}

// NewEtcd returns an allocator that is backed by Etcd and can manage
// persisting the snapshot state of allocation after each allocation is made.
func NewEtcd(alloc Snapshottable, baseKey string, registryKey string, config *storagebackend.Config) *Etcd {
	storage, _ := factory.NewRawStorage(config)

	return &Etcd{
		alloc:       alloc,
		storage:     storage,
		baseKey:     baseKey,
		registryKey: registryKey,
	}
}

// Allocate attempts to allocate the item locally and then in etcd.
func (e *Etcd) Allocate(offset int) (bool, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.storage.Lock(context.TODO())
	defer e.storage.Unlock(context.TODO())

	r, err := e.Get()
	if err != nil {
		return false, err
	}

	e.alloc.Restore(r.Range, r.Data)
	ok, err := e.alloc.Allocate(offset)
	if !ok || err != nil {
		return ok, err
	}

	err = e.update()
	if err != nil {
		return false, err
	}

	return true, nil
}

// AllocateNext attempts to allocate the next item locally and then in etcd.
func (e *Etcd) AllocateNext() (int, bool, error) {
	if err := e.storage.Lock(context.TODO()); err != nil {
		return 0, false, err
	}
	defer e.storage.Unlock(context.TODO())

	r, err := e.Get()
	if err != nil {
		return 0, false, err
	}

	e.alloc.Restore(r.Range, r.Data)
	offset, ok, err := e.alloc.AllocateNext()
	if !ok || err != nil {
		return offset, ok, err
	}

	err = e.update()
	if err != nil {
		return 0, false, err
	}

	return offset, ok, err
}

// Release attempts to release the provided item locally and then in etcd.
func (e *Etcd) Release(item int) error {
	e.storage.Lock(context.TODO())
	defer e.storage.Unlock(context.TODO())

	e.lock.Lock()
	defer e.lock.Unlock()

	r, err := e.Get()
	if err != nil {
		return err
	}
	e.alloc.Restore(r.Range, r.Data)

	err = e.alloc.Release(item)
	if err != nil {
		return err
	}

	err = e.update()
	if err != nil {
		return err
	}

	return nil
}

// ForEach implements loop
func (e *Etcd) ForEach(fn func(int)) {
	e.storage.Lock(context.TODO())
	defer e.storage.Unlock(context.TODO())

	e.lock.Lock()
	defer e.lock.Unlock()

	e.alloc.ForEach(fn)
}

// Has implements allocator.Interface::Has
func (e *Etcd) Has(item int) bool {
	e.storage.Lock(context.TODO())
	defer e.storage.Unlock(context.TODO())

	e.lock.Lock()
	defer e.lock.Unlock()

	return e.alloc.Has(item)
}

// Free implements allocator.Interface::Free
func (e *Etcd) Free() int {
	e.storage.Lock(context.TODO())
	defer e.storage.Unlock(context.TODO())

	e.lock.Lock()
	defer e.lock.Unlock()

	return e.alloc.Free()
}

// Get returns an api.RangeAllocation that represents the current state in
// etcd. If the key does not exist, the object will have an empty ResourceVersion.
func (e *Etcd) Get() (*RangeAllocation, error) {
	existing := &RangeAllocation{}
	if err := e.storage.Get(context.TODO(), e.baseKey, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Init does the storage initialization.
func (e *Etcd) Init() error {
	e.storage.Lock(context.TODO())
	defer e.storage.Unlock(context.TODO())

	e.lock.Lock()
	defer e.lock.Unlock()

	rangeSpec, data := e.alloc.Snapshot()
	r := &RangeAllocation{Range: rangeSpec, Data: data}
	err := e.storage.Create(context.TODO(), e.baseKey, r)
	if err != nil {
		r, err = e.Get()
		if err != nil {
			return err
		}
	}
	return e.alloc.Restore(r.Range, r.Data)
}

// ClearRangeRegistry cleanups the range registry.
func (e *Etcd) ClearRangeRegistry() error {
	return e.storage.Delete(context.TODO(), e.baseKey)
}

func (e *Etcd) update() error {
	rangeSpec, data := e.alloc.Snapshot()
	r := RangeAllocation{Range: rangeSpec, Data: data}
	err := e.storage.Update(context.TODO(), e.baseKey, r)
	return err
}

// Register registers specified ip to id in storage.
func (e *Etcd) Register(ip *net.IP, id string) error {
	return e.storage.Create(context.TODO(), e.registryKey+"/"+id, ip.String())
}

// Deregister 
func (e *Etcd) Deregister(id string) error {
	return e.storage.Delete(context.TODO(), e.registryKey+"/"+id)
}

func (e *Etcd) ClearIPRegistry() error {
	return e.storage.Delete(context.TODO(), e.registryKey)
}

func (e *Etcd) GetIP(id string) (*net.IP, error) {
	ipStr := ""
	err := e.storage.Get(context.TODO(), e.registryKey+"/"+id, &ipStr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(ipStr)
	return &ip, nil
}
