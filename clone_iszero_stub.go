package ref

import (
	"reflect"
)

func CanIsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Complex64, reflect.Complex128:
		return true
	case reflect.Array:
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	case reflect.UnsafePointer:
		return true
	case reflect.String:
		return true
	case reflect.Struct:
		return true
	default:
		return false
	}
}

func CanIsNil(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer:
		return true
	case reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}

func SetZero(to reflect.Value) {
	if CanIsZero(to) {
		to.Set(reflect.Zero(to.Type()))
	}
}

func SetNil(to reflect.Value) {
	if CanIsNil(to) {
		// to.Set(reflect.New(to.Type().Elem()))
		to.Set(reflect.Zero(to.Type()))
	}
}
