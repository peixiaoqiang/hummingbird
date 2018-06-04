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
	"errors"
	"sync"
	"k8s.io/kubernetes/pkg/registry/core/service/allocator"
	"github.com/TalkingData/hummingbird/storage"
	"github.com/TalkingData/hummingbird/storage/storagebackend"
	"github.com/TalkingData/hummingbird/storage/storagebackend/factory"
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

	alloc   allocator.Interface
	storage storage.Interface
	last    string

	baseKey string
}

// Etcd implements allocator.Interface and rangeallocation.RangeRegistry
var _ allocator.Interface = &Etcd{}

// NewEtcd returns an allocator that is backed by Etcd and can manage
// persisting the snapshot state of allocation after each allocation is made.
func NewEtcd(alloc allocator.Interface, baseKey string, config *storagebackend.Config) *Etcd {
	storage, _ := factory.NewRawStorage(config)

	return &Etcd{
		alloc:   alloc,
		storage: storage,
		baseKey: baseKey,
	}
}

// Allocate attempts to allocate the item locally and then in etcd.
func (e *Etcd) Allocate(offset int) (bool, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	ok, err := e.alloc.Allocate(offset)
	if !ok || err != nil {
		return ok, err
	}

	return true, nil
}

// AllocateNext attempts to allocate the next item locally and then in etcd.
func (e *Etcd) AllocateNext() (int, bool, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	offset, ok, err := e.alloc.AllocateNext()
	if !ok || err != nil {
		return offset, ok, err
	}

	return offset, ok, err
}

// Release attempts to release the provided item locally and then in etcd.
func (e *Etcd) Release(item int) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if err := e.alloc.Release(item); err != nil {
		return err
	}

	return nil
}

func (e *Etcd) ForEach(fn func(int)) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.alloc.ForEach(fn)
}

// Implements allocator.Interface::Has
func (e *Etcd) Has(item int) bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.alloc.Has(item)
}

// Implements allocator.Interface::Free
func (e *Etcd) Free() int {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.alloc.Free()
}
