package dino

import (
	"errors"
	"reflect"
	"sync"
)

var (
	ErrNilValue       = errors.New("nil value provided")
	ErrExpectedFunc   = errors.New("expected function")
	ErrExpectedStruct = errors.New("expected struct or pointer to struct")
)

type Dino struct {
	registry *Registry
	mu       sync.Mutex
}

func New() *Dino {
	return &Dino{
		registry: new(Registry),
		mu:       sync.Mutex{},
	}
}

func (d *Dino) WithRegistry(registry *Registry) *Dino {
	d.registry = registry

	return d
}

func (d *Dino) Factory(fn any, tags ...string) error {
	rt := reflect.TypeOf(fn)

	if !isFunction(rt) {
		return ErrExpectedFunc
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(tags) == 0 {
		tags = []string{""}
	}

	for idx := range rt.NumOut() {
		if rt.Out(idx).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		for _, tag := range tags {
			key := Key{
				Tag: tag,
				Ref: rt.Out(idx),
			}

			d.registry.Add(key, reflect.ValueOf(fn))
		}
	}

	return nil
}

func (d *Dino) Singleton(val any, tags ...string) error {
	if val == nil {
		return ErrNilValue
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(tags) == 0 {
		tags = []string{""}
	}

	rt := reflect.TypeOf(val)

	for _, tag := range tags {
		key := Key{
			Tag: tag,
			Ref: rt,
		}

		d.registry.Add(key, reflect.ValueOf(val))
	}

	return nil
}

func (d *Dino) Inject(target any) error {
	rt := reflect.TypeOf(target)

	if !isStruct(rt) && !isPointerToStruct(rt) {
		return ErrExpectedStruct
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	resolver := NewResolver(d.registry)
	resolver.Inject(reflect.ValueOf(target))

	if len(resolver.errors) == 0 {
		return nil
	}

	return resolver
}

func (d *Dino) Execute(fn any) error {
	rt := reflect.TypeOf(fn)

	if !isFunction(rt) {
		return ErrExpectedFunc
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	resolver := NewResolver(d.registry)
	resolver.Execute(reflect.ValueOf(fn))

	if len(resolver.errors) == 0 {
		return nil
	}

	return resolver
}
