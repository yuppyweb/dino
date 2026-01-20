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

type Injector struct {
	registry Registry
	stack    map[RegistryKey]int32
}

func NewInjector(registry Registry) *Injector {
	if registry == nil {
		registry = new(SyncMapRegistry)
	}

	return &Injector{
		registry: registry,
		stack:    make(map[RegistryKey]int32),
	}
}

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

		if !errors.Is(err, ErrValueNotFound) {
			return fmt.Errorf("resolve field %s: %w", fieldStruct.Name, err)
		}

		val = i.Create(fieldType)

		if err := i.Inject(val); err != nil {
			if !errors.Is(err, ErrExpectedStruct) {
				return fmt.Errorf("inject field %s: %w", fieldStruct.Name, err)
			}
		}

		field.Set(val)
	}

	return nil
}

func (i *Injector) Invoke(rv reflect.Value) ([]reflect.Value, error) {
	rt := rv.Type()

	if !isFunction(rt) {
		return nil, fmt.Errorf("%w: got %s", ErrExpectedFunction, rt.Kind())
	}

	args, err := i.Prepare(rt)
	if err != nil {
		return nil, fmt.Errorf("prepare function execution arguments: %w", err)
	}

	return rv.Call(args), nil
}

func (i *Injector) Resolve(key RegistryKey) (reflect.Value, error) {
	rv, err := i.registry.Find(key)
	if err != nil {
		return rv, fmt.Errorf("resolve type %s with tag '%s': %w", key.Type, key.Tag, err)
	}

	resVal := reflect.Zero(key.Type)

	// Detect circular dependencies
	if i.stack[key] > 0 {
		return resVal, fmt.Errorf(
			"%w: type %s with tag '%s'",
			ErrCircularDependency,
			key.Type,
			key.Tag,
		)
	}

	// Mark as being resolved
	i.stack[key]++

	defer func() {
		// Unmark after resolution
		i.stack[key]--
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

		values := rv.Call(args)

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

func (i *Injector) Prepare(fn reflect.Type) ([]reflect.Value, error) {
	if !isFunction(fn) {
		return nil, fmt.Errorf("%w: got %s", ErrExpectedFunction, fn.Kind())
	}

	// Prepare arguments
	num := fn.NumIn()
	arg := make([]reflect.Value, num)

	for idx := range num {
		rt := fn.In(idx)

		key := RegistryKey{
			Tag:  "",
			Type: rt,
		}

		rv, err := i.Resolve(key)
		if err == nil {
			arg[idx] = rv

			continue
		}

		if !errors.Is(err, ErrValueNotFound) {
			return nil, fmt.Errorf("resolve argument of type %s: %w", rt, err)
		}

		rv = i.Create(rt)

		if err := i.Inject(rv); err != nil {
			if !errors.Is(err, ErrExpectedStruct) {
				return nil, fmt.Errorf("inject argument of type %s: %w", rt, err)
			}
		}

		arg[idx] = rv
	}

	return arg, nil
}

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
