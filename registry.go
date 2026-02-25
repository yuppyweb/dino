package dino

import (
	"errors"
	"reflect"
	"sync"
)

var (
	ErrKeyTypeNil    = errors.New("registry key type cannot be nil")
	ErrValueNotFound = errors.New("value not found in registry")
	ErrInvalidValue  = errors.New("registry invalid value")
)

// Defines the interface for a dependency registry.
type Registry interface {
	Register(key RegistryKey, rv reflect.Value) error
	Find(key RegistryKey) (reflect.Value, error)
}

// Represents a unique key for a dependency in the registry, consisting of a tag and a type.
type RegistryKey struct {
	// An optional tag to differentiate between multiple values of the same type.
	Tag string

	// The type of the value being registered.
	Type reflect.Type
}

// Is a thread-safe implementation of the Registry interface using sync.Map.
type SyncMapRegistry struct {
	// Internal storage for the registry using sync.Map for concurrent access.
	sm sync.Map
}

// Registers a value in the registry with the specified key.
func (r *SyncMapRegistry) Register(key RegistryKey, rv reflect.Value) error {
	if key.Type == nil {
		return ErrKeyTypeNil
	}

	if !rv.IsValid() {
		return ErrInvalidValue
	}

	r.sm.Store(key, rv)

	return nil
}

// Finds a value in the registry based on the specified key.
func (r *SyncMapRegistry) Find(key RegistryKey) (reflect.Value, error) {
	if key.Type == nil {
		return reflect.Value{}, ErrKeyTypeNil
	}

	value, ok := r.sm.Load(key)
	if !ok {
		return reflect.Zero(key.Type), ErrValueNotFound
	}

	rv, ok := value.(reflect.Value)
	if !ok {
		return reflect.Zero(key.Type), ErrInvalidValue
	}

	return rv, nil
}

// Ensure SyncMapRegistry implements the Registry interface.
var _ Registry = (*SyncMapRegistry)(nil)
