package dino

import (
	"fmt"
	"reflect"
	"sync"
)

type DependencyInjector interface {
	Bind(rt reflect.Type, rv reflect.Value, tags ...string) error
	Inject(rv reflect.Value) error
	Invoke(rv reflect.Value) error
}

type Dino struct {
	di DependencyInjector
	mu sync.Mutex
}

func New() *Dino {
	return &Dino{
		di: NewInjector(),
		mu: sync.Mutex{},
	}
}

func (d *Dino) WithDI(di DependencyInjector) *Dino {
	d.di = di

	return d
}

func (d *Dino) Factory(fn any, tags ...string) error {
	rt := reflect.TypeOf(fn)

	if !isFunction(rt) {
		return fmt.Errorf("%w: got %s", ErrExpectedFunction, rt.Kind())
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for idx := range rt.NumOut() {
		outType := rt.Out(idx)
		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		if err := d.di.Bind(outType, reflect.ValueOf(fn), tags...); err != nil {
			return fmt.Errorf("failed to bind factory function output: %w", err)
		}
	}

	return nil
}

func (d *Dino) Singleton(val any, tags ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.di.Bind(reflect.TypeOf(val), reflect.ValueOf(val), tags...); err != nil {
		return fmt.Errorf("failed to bind singleton: %w", err)
	}

	return nil
}

func (d *Dino) Inject(target any) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.di.Inject(reflect.ValueOf(target)); err != nil {
		return fmt.Errorf("failed to inject dependencies: %w", err)
	}

	return nil
}

func (d *Dino) Invoke(fn any) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.di.Invoke(reflect.ValueOf(fn)); err != nil {
		return fmt.Errorf("failed to invoke function: %w", err)
	}

	return nil
}
