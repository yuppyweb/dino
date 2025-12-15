package dino

import (
	"reflect"
)

func isStruct(rt reflect.Type) bool {
	return rt.Kind() == reflect.Struct
}

func isPointerToStruct(rt reflect.Type) bool {
	return rt.Kind() == reflect.Pointer && isStruct(rt.Elem())
}

func isFunction(rt reflect.Type) bool {
	return rt.Kind() == reflect.Func
}

func isNil(rv reflect.Value) bool {
	switch rv.Type().Kind() {
	case reflect.Invalid:
		return true

	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()

	default:
		return false
	}
}
