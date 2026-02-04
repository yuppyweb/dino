// Package dino provides a simple dependency injection container for Go.
// It allows registering factories and singletons, injecting dependencies into structs,
// and invoking functions with automatic dependency resolution.
package dino

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var ErrInvalidInputValue = errors.New("invalid input value")

// Is the main struct for the Dino dependency injection container.
type Dino struct {
	// registry holds the registered dependencies.
	registry Registry

	// mutex is used to ensure thread-safe operations.
	mutex sync.Mutex
}

// Creates a new instance of the Dino dependency injection container.
func New() *Dino {
	return &Dino{
		registry: new(SyncMapRegistry),
		mutex:    sync.Mutex{},
	}
}

// Sets a custom registry for the Dino container.
func (d *Dino) WithRegistry(registry Registry) *Dino {
	d.registry = registry

	return d
}

// Registers a factory function that produces instances of dependencies.
func (d *Dino) Factory(fn any, tags ...string) error {
	rv := reflect.ValueOf(fn)

	if isNil(rv) {
		return fmt.Errorf("%w: factory function cannot be nil", ErrInvalidInputValue)
	}

	rt := rv.Type()

	if !isFunction(rt) {
		return fmt.Errorf(
			"%w: factory expected a function, got %v",
			ErrInvalidInputValue,
			rt.Kind(),
		)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	injector := NewInjector(d.registry)

	for idx := range rt.NumOut() {
		outType := rt.Out(idx)
		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		if err := injector.Bind(outType, reflect.ValueOf(fn), tags...); err != nil {
			return fmt.Errorf("failed to bind factory function output: %w", err)
		}
	}

	return nil
}

// Registers a singleton instance of a dependency.
func (d *Dino) Singleton(val any, tags ...string) error {
	rv := reflect.ValueOf(val)

	if isNil(rv) {
		return fmt.Errorf("%w: singleton value cannot be nil", ErrInvalidInputValue)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	injector := NewInjector(d.registry)

	if err := injector.Bind(reflect.TypeOf(val), rv, tags...); err != nil {
		return fmt.Errorf("failed to bind singleton: %w", err)
	}

	return nil
}

// Injects dependencies into the provided target struct.
func (d *Dino) Inject(target any) error {
	rv := reflect.ValueOf(target)

	if isNil(rv) {
		return fmt.Errorf("%w: inject target cannot be nil", ErrInvalidInputValue)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	injector := NewInjector(d.registry)

	if err := injector.Inject(rv); err != nil {
		return fmt.Errorf("failed to inject dependencies: %w", err)
	}

	return nil
}

// Invokes a function with automatic dependency resolution.
func (d *Dino) Invoke(fn any) ([]any, error) {
	rv := reflect.ValueOf(fn)

	if isNil(rv) {
		return nil, fmt.Errorf("%w: function to invoke cannot be nil", ErrInvalidInputValue)
	}

	rt := rv.Type()

	if !isFunction(rt) {
		return nil, fmt.Errorf(
			"%w: invoke expected a function, got %v",
			ErrInvalidInputValue,
			rt.Kind(),
		)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	injector := NewInjector(d.registry)

	values, err := injector.Invoke(rv)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}

	results := make([]any, len(values))

	for idx, val := range values {
		if !val.CanInterface() {
			results[idx] = nil

			continue
		}

		results[idx] = val.Interface()
	}

	return results, nil
}
