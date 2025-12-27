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
	registry *Registry
	stack    map[Key]int32
}

func NewInjector() *Injector {
	return &Injector{
		registry: new(Registry),
		stack:    make(map[Key]int32),
	}
}

func (i *Injector) WithRegistry(registry *Registry) *Injector {
	i.registry = registry

	return i
}

func (i *Injector) Bind(rt reflect.Type, rv reflect.Value, tags ...string) error {
	if len(tags) == 0 {
		tags = []string{""}
	}

	for _, tag := range tags {
		key := Key{
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

		key := Key{
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

func (i *Injector) Invoke(rv reflect.Value) error {
	rt := rv.Type()

	if !isFunction(rt) {
		return fmt.Errorf("%w: got %s", ErrExpectedFunction, rt.Kind())
	}

	args, err := i.Prepare(rt)
	if err != nil {
		return fmt.Errorf("prepare function execution arguments: %w", err)
	}

	values := rv.Call(args)

	for _, val := range values {
		if err := asError(val); err != nil {
			return fmt.Errorf("function execution returned error: %w", err)
		}

		// Skip nil values
		if isNil(val) {
			continue
		}

		key := Key{
			Tag:  "",
			Type: val.Type(),
		}

		if err := i.registry.Register(key, val); err != nil {
			return fmt.Errorf(
				"failed to add value of type %s to registry: %w",
				val.Type().String(),
				err,
			)
		}
	}

	return nil
}

func (i *Injector) Resolve(key Key) (reflect.Value, error) {
	resVal := reflect.Zero(key.Type)

	rv, err := i.registry.Find(key)
	if err != nil {
		return resVal, fmt.Errorf("resolve type %s with tag '%s': %w", key.Type, key.Tag, err)
	}

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
				key.Type.String(),
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

			valKey := Key{
				Tag:  key.Tag,
				Type: val.Type(),
			}

			if err := i.registry.Register(valKey, val); err != nil {
				return resVal, fmt.Errorf(
					"add value of type %s with tag '%s' to registry: %w",
					val.Type().String(),
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

		key := Key{
			Tag:  "",
			Type: rt,
		}

		rv, err := i.Resolve(key)
		if err == nil {
			arg[idx] = rv

			continue
		}

		if !errors.Is(err, ErrValueNotFound) {
			return nil, fmt.Errorf("resolve argument of type %s: %w", rt.String(), err)
		}

		rv = i.Create(rt)

		if err := i.Inject(rv); err != nil {
			if !errors.Is(err, ErrExpectedStruct) {
				return nil, fmt.Errorf("inject argument of type %s: %w", rt.String(), err)
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
