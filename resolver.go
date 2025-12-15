package dino

import (
	"errors"
	"fmt"
	"reflect"
)

type Resolver struct {
	registry *Registry
	stack    map[Key]int32
	errors   []error
}

func NewResolver(registry *Registry) *Resolver {
	return &Resolver{
		registry: registry,
		stack:    make(map[Key]int32),
		errors:   make([]error, 0),
	}
}

func (r *Resolver) Inject(rv reflect.Value) {
	if !isStruct(rv.Type()) && !isPointerToStruct(rv.Type()) {
		return
	}

	// If pointer, get the element
	rv = reflect.Indirect(rv)
	rt := rv.Type()

	// Iterate over fields
	for idx := range rv.NumField() {
		field := rv.Field(idx)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		fieldType := field.Type()

		// Get tag value for "inject"
		tag := rt.Field(idx).Tag.Get("inject")

		key := Key{
			Tag: tag,
			Ref: fieldType,
		}

		if val, ok := r.Resolve(key); ok {
			field.Set(val)

			continue
		}

		val := r.Create(fieldType)

		if isStruct(fieldType) || isPointerToStruct(fieldType) {
			r.Inject(val)
		}

		field.Set(val)
	}
}

func (r *Resolver) Execute(rv reflect.Value) {
	rt := rv.Type()

	if !isFunction(rt) {
		return
	}

	// Call the function with prepared arguments
	values := rv.Call(r.Prepare(rt))

	for _, val := range values {
		// Check if value is error
		if val.CanInterface() {
			if err, ok := val.Interface().(error); ok {
				if err != nil {
					r.errors = append(r.errors, err)
				}

				continue
			}
		}

		// Skip nil values
		if isNil(val) {
			continue
		}

		key := Key{
			Tag: "",
			Ref: val.Type(),
		}

		r.registry.Add(key, val)
	}
}

func (r *Resolver) Resolve(key Key) (reflect.Value, bool) {
	rv, ok := r.registry.Get(key)
	if !ok {
		return rv, false
	}

	// Detect circular dependencies
	if r.stack[key] > 0 {
		r.errors = append(
			r.errors,
			fmt.Errorf(
				"circular dependency detected for type %s with tag '%s'",
				key.Ref.String(),
				key.Tag,
			),
		)

		return rv, false
	}

	// Mark as being resolved
	r.stack[key]++
	defer func() {
		// Unmark after resolution
		r.stack[key]--
	}()

	rt := rv.Type()

	// If the registered value is a factory function, call it to get the actual value
	if isFunction(rt) && rt != key.Ref {
		// Call the factory function
		values := rv.Call(r.Prepare(rt))

		for _, val := range values {
			// Check if value is error
			if val.CanInterface() {
				if err, ok := val.Interface().(error); ok {
					if err != nil {
						r.errors = append(r.errors, err)
					}

					continue
				}
			}

			// Skip nil values
			if isNil(val) {
				continue
			}

			valKey := Key{
				Tag: key.Tag,
				Ref: val.Type(),
			}

			r.registry.Add(valKey, val)

			// Return matching type
			if val.Type() == key.Ref {
				rv = val
			}
		}
	}

	return rv, rv.Type() == key.Ref
}

func (r *Resolver) Prepare(fn reflect.Type) []reflect.Value {
	if !isFunction(fn) {
		return nil
	}

	// Prepare arguments
	num := fn.NumIn()
	arg := make([]reflect.Value, num)

	for idx := range num {
		rt := fn.In(idx)

		key := Key{
			Tag: "",
			Ref: rt,
		}

		if rv, ok := r.Resolve(key); ok {
			arg[idx] = rv

			continue
		}

		rv := r.Create(rt)

		// If struct or pointer to struct, inject dependencies
		if isStruct(rt) || isPointerToStruct(rt) {
			r.Inject(rv)
		}

		arg[idx] = rv
	}

	return arg
}

func (r *Resolver) Create(rt reflect.Type) reflect.Value {
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
		return reflect.MakeFunc(rt, func(in []reflect.Value) []reflect.Value {
			numOut := rt.NumOut()
			values := make([]reflect.Value, numOut)

			for idx := range numOut {
				// Recursively create each return value
				values[idx] = r.Create(rt.Out(idx))
			}

			return values
		})

	default:
		return reflect.Zero(rt)
	}
}

func (r *Resolver) Error() string {
	if len(r.errors) == 0 {
		return ""
	}

	return errors.Join(r.errors...).Error()
}

func (r *Resolver) Unwrap() []error {
	return r.errors
}
