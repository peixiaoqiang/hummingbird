package storage

import (
	"context"
)

type Object interface {
}

// Interface offers a common interface for object marshaling/unmarshaling operations and
// hides all the storage-related operations behind it.
type Interface interface {
	// Create adds a new object at a key unless it already exists. 'ttl' is time-to-live
	// in seconds (0 means forever). If no error is returned and out is not nil, out will be
	// set to the read value from database.
	Create(ctx context.Context, key string, obj Object) error

	Update(ctx context.Context, key string, obj Object) error

	CreateOrUpdate(ctx context.Context, key string, obj Object) error

	// Delete removes the specified key and returns the value that existed at that spot.
	// If key didn't exist, it will return NotFound storage error.
	Delete(ctx context.Context, key string) error

	// Get unmarshals json found at key into objPtr. On a not found error, will either
	// return a zero object of the requested type, or an error, depending on ignoreNotFound.
	// Treats empty responses and nil response nodes exactly like a not found error.
	// The returned contents may be delayed, but it is guaranteed that they will
	// be have at least 'resourceVersion'.
	Get(ctx context.Context, key string, objPtr Object) error

	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}
