package dino

import (
	"errors"
	"reflect"
	"sync"
)

var (
	ErrKeyTypeNil    = errors.New("key type cannot be nil")
	ErrValueNotFound = errors.New("value not found in registry")
	ErrInvalidValue  = errors.New("invalid value")
)

type Key struct {
	Tag  string
	Type reflect.Type
}

type Registry struct {
	sm sync.Map
}

func (r *Registry) Register(key Key, rv reflect.Value) error {
	if key.Type == nil {
		return ErrKeyTypeNil
	}

	if !rv.IsValid() {
		return ErrInvalidValue
	}

	r.sm.Store(key, rv)

	return nil
}

func (r *Registry) Find(key Key) (reflect.Value, error) {
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
