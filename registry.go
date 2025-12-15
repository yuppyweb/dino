package dino

import (
	"reflect"
	"sync"
)

type Key struct {
	Tag string
	Ref reflect.Type
}

type Registry struct {
	sm sync.Map
}

func (d *Registry) Add(key Key, rv reflect.Value) {
	d.sm.Store(key, rv)
}

func (d *Registry) Get(key Key) (reflect.Value, bool) {
	value, ok := d.sm.Load(key)
	if !ok {
		return reflect.Zero(key.Ref), false
	}

	return value.(reflect.Value), true
}
