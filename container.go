package dino

import (
	"fmt"
	"reflect"
	"sync"
)

type storageKey struct {
	tag string
	key reflect.Type
}

type Container struct {
	storage *sync.Map
}

func New() *Container {
	return &Container{
		storage: &sync.Map{},
	}
}

// Factory registers a factory function with an optional tag.
func (c *Container) Factory(tag string, fn any) error {
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("expected function, got %s", fnType.Kind()) //nolint:err113
	}

	for index := range fnType.NumOut() {
		key := storageKey{
			tag: tag,
			key: fnType.Out(index),
		}

		c.storage.Store(key, fn)
	}

	return nil
}

// Singleton registers a singleton factory function.
func (c *Container) Singleton(fn any) error {
	return c.Factory("", fn)
}

// Inject populates the fields of the target struct with dependencies from the container.
func (c *Container) Inject(target any) error {
	return c.reflectInject(reflect.ValueOf(target))
}

func (c *Container) reflectInject(value reflect.Value) error {
	value = reflect.Indirect(value)

	if value.Kind() != reflect.Struct {
		return nil
	}

	typeOf := value.Type()

	for index := range value.NumField() {
		field := value.Field(index)

		if !field.CanSet() {
			continue
		}

		if err := c.reflectInject(field); err != nil {
			return err
		}

		tag := typeOf.Field(index).Tag.Get("inject")

		val, err := c.resolve(field.Type(), tag)
		if err != nil {
			return err
		}

		field.Set(val)
	}

	return nil
}

func (c *Container) resolve(typeOf reflect.Type, tag string) (reflect.Value, error) {
	key := storageKey{
		tag: tag,
		key: typeOf,
	}

	value, ok := c.storage.Load(key)
	if !ok {
		return reflect.Zero(typeOf), nil
	}

	valueOf := reflect.ValueOf(value)

	if valueOf.Kind() == reflect.Func {
		args, err := c.resolveFunc(valueOf.Type())
		if err != nil {
			return reflect.Zero(typeOf), err
		}

		values := valueOf.Call(args)

		if err := c.detectError(values); err != nil {
			return reflect.Zero(typeOf), err
		}

		c.swap(values, tag)

		for _, val := range values {
			if val.Type() == typeOf {
				return val, nil
			}
		}

		return reflect.Zero(typeOf), fmt.Errorf("no value of type %s", typeOf) //nolint:err113
	}

	return valueOf, nil
}

func (c *Container) resolveFunc(fn reflect.Type) ([]reflect.Value, error) {
	args := make([]reflect.Value, fn.NumIn())

	for index := range fn.NumIn() {
		argType := fn.In(index)

		argValue, err := c.resolve(argType, "")
		if err != nil {
			return nil, err
		}

		args[index] = argValue
	}

	return args, nil
}

func (c *Container) detectError(values []reflect.Value) error {
	for _, val := range values {
		if !val.CanInterface() {
			continue
		}

		if err, ok := val.Interface().(error); ok && err != nil {
			return fmt.Errorf("error detected in return values: %w", err)
		}
	}

	return nil
}

func (c *Container) swap(values []reflect.Value, tag string) {
	for _, val := range values {
		if val.CanInterface() {
			if _, ok := val.Interface().(error); ok {
				continue
			}
		}

		key := storageKey{
			tag: tag,
			key: val.Type(),
		}

		c.storage.Store(key, val)
	}
}
