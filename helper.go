package dino

import (
	"reflect"
)

// Helper functions for type checking and error handling.
func isStruct(rt reflect.Type) bool {
	return rt.Kind() == reflect.Struct
}

// Checks if the given type is a pointer to a struct.
func isPointerToStruct(rt reflect.Type) bool {
	return rt.Kind() == reflect.Pointer && isStruct(rt.Elem())
}

// Checks if the given type is a function.
func isFunction(rt reflect.Type) bool {
	return rt.Kind() == reflect.Func
}

// Checks if the given reflect.Value is nil or invalid.
func isNil(rv reflect.Value) bool {
	if !rv.IsValid() {
		return true
	}

	switch rv.Type().Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()

	default:
		return false
	}
}

// Converts a reflect.Value to an error if it implements the error interface and is not nil.
func asError(rv reflect.Value) error {
	if isNil(rv) || !rv.CanInterface() {
		return nil
	}

	if err, ok := rv.Interface().(error); ok && err != nil {
		return err
	}

	return nil
}
