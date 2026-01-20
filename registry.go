package dino

import (
	"errors"
	"reflect"
	"sync"
)

var (
	ErrKeyTypeNil    = errors.New("for registry key type cannot be nil")
	ErrValueNotFound = errors.New("value not found in registry")
	ErrInvalidValue  = errors.New("registry invalid value")
)

type Registry interface {
	Register(key RegistryKey, rv reflect.Value) error
	Find(key RegistryKey) (reflect.Value, error)
}

type RegistryKey struct {
	Tag  string
	Type reflect.Type
}

type SyncMapRegistry struct {
	sm sync.Map
}

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

var _ Registry = (*SyncMapRegistry)(nil)
