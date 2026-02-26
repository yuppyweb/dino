package dino

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrExpectedStruct     = errors.New("expected struct or pointer to struct")
	ErrExpectedFunction   = errors.New("expected function")
	ErrCircularDependency = errors.New("circular dependency detected")
)

// Injector is responsible for managing dependencies, injecting values into structs,
// and invoking functions with resolved arguments.
type Injector struct {
	registry Registry
	stack    map[RegistryKey]struct{}
}

// NewInjector creates a new Injector with the provided registry.
// If no registry is provided, it uses a default SyncMapRegistry.
func NewInjector(registry Registry) *Injector {
	if registry == nil {
		registry = new(SyncMapRegistry)
	}

	return &Injector{
		registry: registry,
		stack:    make(map[RegistryKey]struct{}),
	}
}

// Bind registers a value in the registry for the specified type and optional tags.
func (i *Injector) Bind(rt reflect.Type, rv reflect.Value, tags ...string) error {
	if len(tags) == 0 {
		tags = []string{""}
	}

	for _, tag := range tags {
		key := RegistryKey{
			Tag:  tag,
			Type: rt,
		}

		if err := i.registry.Register(key, rv); err != nil {
			return fmt.Errorf("bind value to registry: %w", err)
		}
	}

	return nil
}

// Inject resolves and sets dependencies on the provided struct value based on "inject" tags and registered values.
func (i *Injector) Inject(rv reflect.Value) error {
	rt := rv.Type()

	if isPointerToStruct(rt) {
		// If pointer to struct, get struct value
		rv = reflect.Indirect(rv)
		rt = rv.Type()
	}

	if !isStruct(rt) {
		return fmt.Errorf("%w: got %s", ErrExpectedStruct, rt.Kind())
	}

	// Iterate over fields
	for idx := range rv.NumField() {
		field := rv.Field(idx)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		fieldType := field.Type()
		fieldStruct := rt.Field(idx)

		// Get tag value for "inject"
		tag := fieldStruct.Tag.Get("inject")

		key := RegistryKey{
			Tag:  tag,
			Type: fieldType,
		}

		val, err := i.Resolve(key)
		if err == nil {
			field.Set(val)

			continue
		}

		// If the error is not ErrValueNotFound, return it
		if !errors.Is(err, ErrValueNotFound) {
			return fmt.Errorf("resolve field %s: %w", fieldStruct.Name, err)
		}

		// If value not found, create a new instance and inject it
		val = i.Create(fieldType)

		// If the field is a struct or pointer to struct, inject dependencies into it
		if err := i.Inject(val); err != nil {
			if !errors.Is(err, ErrExpectedStruct) {
				return fmt.Errorf("inject field %s: %w", fieldStruct.Name, err)
			}
		}

		field.Set(val)
	}

	return nil
}

// Invoke calls a function with arguments resolved from the registry. The function must be passed as a reflect.Value.
func (i *Injector) Invoke(rv reflect.Value) ([]reflect.Value, error) {
	rt := rv.Type()

	if !isFunction(rt) {
		return nil, fmt.Errorf("%w: got %s", ErrExpectedFunction, rt.Kind())
	}

	// Prepare arguments for the function call
	args, err := i.Prepare(rt)
	if err != nil {
		return nil, fmt.Errorf("prepare function execution arguments: %w", err)
	}

	return rv.Call(args), nil
}

// Resolve looks up a value from the registry based on the provided key.
// If the registered value is a factory function, it calls the function to get the actual value.
func (i *Injector) Resolve(key RegistryKey) (reflect.Value, error) {
	rv, err := i.registry.Find(key)
	if err != nil {
		return rv, fmt.Errorf("resolve type %s with tag '%s': %w", key.Type, key.Tag, err)
	}

	resVal := reflect.Zero(key.Type)

	// Detect circular dependencies
	if _, exists := i.stack[key]; exists {
		return resVal, fmt.Errorf(
			"%w: type %s with tag '%s'",
			ErrCircularDependency,
			key.Type,
			key.Tag,
		)
	}

	// Mark as being resolved
	i.stack[key] = struct{}{}

	defer func() {
		// Unmark after resolution
		delete(i.stack, key)
	}()

	rt := rv.Type()

	// If the registered value is a factory function, call it to get the actual value
	if isFunction(rt) && rt != key.Type {
		args, err := i.Prepare(rt)
		if err != nil {
			return resVal, fmt.Errorf(
				"prepare factory function arguments of type %s with tag '%s': %w",
				key.Type,
				key.Tag,
				err,
			)
		}

		// Call the factory function
		values := rv.Call(args)

		// Process the returned values from the factory function
		for _, val := range values {
			if err := asError(val); err != nil {
				return resVal, fmt.Errorf(
					"factory function for type %s with tag '%s' returned error: %w",
					key.Type,
					key.Tag,
					err,
				)
			}

			// Skip nil values
			if isNil(val) {
				continue
			}

			// Bind the returned value to the registry for future resolutions
			if err := i.Bind(val.Type(), val, key.Tag); err != nil {
				return resVal, fmt.Errorf(
					"bind factory function return value of type %s with tag '%s': %w",
					val.Type(),
					key.Tag,
					err,
				)
			}

			// Return matching type
			if val.Type() == key.Type {
				resVal = val
			}
		}

		return resVal, nil
	}

	return rv, nil
}

// Prepare builds the arguments for a function call by resolving them from the registry
// or creating new instances if not found.
func (i *Injector) Prepare(fn reflect.Type) ([]reflect.Value, error) {
	if !isFunction(fn) {
		return nil, fmt.Errorf("%w: got %s", ErrExpectedFunction, fn.Kind())
	}

	// Prepare arguments
	num := fn.NumIn()
	arg := make([]reflect.Value, num)

	// Iterate over function parameters
	for idx := range num {
		rt := fn.In(idx)

		key := RegistryKey{
			Tag:  "",
			Type: rt,
		}

		// Try to resolve the argument from the registry
		rv, err := i.Resolve(key)
		if err == nil {
			arg[idx] = rv

			continue
		}

		// If the error is not ErrValueNotFound, return it
		if !errors.Is(err, ErrValueNotFound) {
			return nil, fmt.Errorf("resolve argument of type %s: %w", rt, err)
		}

		// If value not found, create a new instance and inject it
		rv = i.Create(rt)

		// If the argument is a struct or pointer to struct, inject dependencies into it
		if err := i.Inject(rv); err != nil {
			if !errors.Is(err, ErrExpectedStruct) {
				return nil, fmt.Errorf("inject argument of type %s: %w", rt, err)
			}
		}

		arg[idx] = rv
	}

	return arg, nil
}

// Create returns a new instance of the specified type.
// For complex types like slices, maps, channels, pointers, and functions,
// it creates appropriate zero values or factory functions.
func (i *Injector) Create(rt reflect.Type) reflect.Value {
	switch rt.Kind() {
	case reflect.Slice:
		return reflect.MakeSlice(rt, 0, 0)

	case reflect.Map:
		return reflect.MakeMap(rt)

	case reflect.Chan:
		return reflect.MakeChan(rt, 0)

	case reflect.Pointer:
		return reflect.New(rt.Elem())

	case reflect.Func:
		// Create a function that resolves its return values
		return reflect.MakeFunc(rt, func([]reflect.Value) []reflect.Value {
			numOut := rt.NumOut()
			values := make([]reflect.Value, numOut)

			for idx := range numOut {
				// Recursively create each return value
				values[idx] = i.Create(rt.Out(idx))
			}

			return values
		})

	default:
		return reflect.Zero(rt)
	}
}
