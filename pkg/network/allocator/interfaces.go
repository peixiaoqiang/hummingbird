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

package allocator

import "net"

// Interface manages the allocation of items out of a range. Interface
// should be threadsafe.
type Interface interface {
	Allocate(int) (bool, error)
	AllocateNext() (int, bool, error)
	Release(int) error
	ForEach(func(int))

	// For testing
	Has(int) bool

	// For testing
	Free() int
}

// Snapshottable is an Interface that can be snapshotted and restored. Snapshottable
// should be threadsafe.
type Snapshottable interface {
	Interface
	Snapshot() (string, []byte)
	Restore(string, []byte) error
}

// RangeAllocation is an opaque API object (not exposed to end users) that can be persisted to record
// the global allocation state of the cluster. The schema of Range and Data generic, in that Range
// should be a string representation of the inputs to a range (for instance, for IP allocation it
// might be a CIDR) and Data is an opaque blob understood by an allocator which is typically a
// binary range.  Consumers should use annotations to record additional information (schema version,
// data encoding hints). A range allocation should *ALWAYS* be recreatable at any time by observation
// of the cluster, thus the object is less strongly typed than most.
type RangeAllocation struct {
	Range string
	// A byte array representing the serialized state of a range allocation. Additional clarifiers on
	// the type or format of data should be represented with annotations. For IP allocations, this is
	// represented as a bit array starting at the base IP of the CIDR in Range, with each bit representing
	// a single allocated address (the fifth bit on CIDR 10.0.0.0/8 is 10.0.0.4).
	Data []byte
}

// RangeRegistry is a registry that can retrieve or persist a RangeAllocation object.
type RangeRegistry interface {
	Get() (*RangeAllocation, error)
	Init() error
	// For test case.
	ClearRangeRegistry() error
}

// Factory represents a allocator factory to create allocator.
type Factory func(max int, rangeSpec string) Interface

// IP represents a ip.
type IP struct {
	IP      *net.IPNet
	Gateway net.IP
	Routes  []*Route
}

// Route represents a route rule.
type Route struct {
	Dst *net.IPNet
	GW  net.IP
}

// IPRegistry represents a ip registry to manipulate ip.
type IPRegistry interface {
	Register(*net.IP, string) error
	Deregister(string) error
	// For test case
	ClearIPRegistry() error
	GetIP(id string) (*net.IP, error)
}
