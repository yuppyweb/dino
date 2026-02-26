package dino

import (
	"reflect"
)

// isStruct reports whether rt is a struct type.
func isStruct(rt reflect.Type) bool {
	return rt.Kind() == reflect.Struct
}

// isPointerToStruct reports whether rt is a pointer to a struct type.
func isPointerToStruct(rt reflect.Type) bool {
	return rt.Kind() == reflect.Pointer && isStruct(rt.Elem())
}

// isFunction reports whether rt is a function type.
func isFunction(rt reflect.Type) bool {
	return rt.Kind() == reflect.Func
}

// isNil reports whether rv is nil or invalid.
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

// asError extracts an error from rv if it implements the error interface and is not nil.
func asError(rv reflect.Value) error {
	if isNil(rv) || !rv.CanInterface() {
		return nil
	}

	if err, ok := rv.Interface().(error); ok && err != nil {
		return err
	}

	return nil
}
