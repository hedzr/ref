// Copyright © 2020 Hedzr Yeh.

// +build !go1.13

package ref

import (
	"reflect"
)

// IsZero reports whether v is the zero value for its type.
// It panics if the argument is invalid.
func IsZero(v reflect.Value) bool {
	//switch v.Kind() {
	//case reflect.Bool:
	//	return !v.Bool()
	//case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	//	return v.Int() == 0
	//case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	//	return v.Uint() == 0
	//case reflect.Float32, reflect.Float64:
	//	return math.Float64bits(v.Float()) == 0
	//case reflect.Complex64, reflect.Complex128:
	//	c := v.Complex()
	//	return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	//case reflect.Array:
	//	return isZeroArray(v)
	//case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
	//	return v.IsNil()
	//case reflect.UnsafePointer:
	//	return isNil(v)
	//case reflect.String:
	//	return v.Len() == 0
	//case reflect.Struct:
	//	return isZeroStruct(v)
	//default:
	//	// This should never happens, but will act as a safeguard for
	//	// later, as a default value doesn't makes sense here.
	//	panic(fmt.Sprintf("reflect.Value.IsZero, kind=%b", v.Kind()))
	//}
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !v.Index(i).IsZero() {
				return false
			}
		}
		return true
		// return isZeroArray(v)
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return IsNil(v)
	case reflect.UnsafePointer:
		return IsNil(v)
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsZero() {
				return false
			}
		}
		return true
		// return isZeroStruct(v)
	default:
		// This should never happens, but will act as a safeguard for
		// later, as a default value doesn't makes sense here.
		// panic(fmt.Sprintf("reflect.Value.IsZero, kind=%b", v.Kind()))
		return false
	}
}

// IsNil reports whether its argument v is nil. The argument must be
// a chan, func, interface, map, pointer, or slice value; if it is
// not, IsNil panics. Note that IsNil is not always equivalent to a
// regular comparison with nil in Go. For example, if v was created
// by calling ValueOf with an uninitialized interface variable i,
// i==nil will be true but v.IsNil will panic as v will be the zero
// Value.
func IsNil(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer:
		if v.flag&flagMethod != 0 {
			return false
		}
		ptr := v.ptr
		if v.flag&flagIndir != 0 {
			ptr = *(*unsafe.Pointer)(ptr)
		}
		return ptr == nil
	case reflect.Interface, reflect.Slice:
		// Both interface and slice are nil if first word is 0.
		// Both are always bigger than a word; assume flagIndir.
		return *(*unsafe.Pointer)(v.ptr) == nil
	}
	return false
}

func isZeroArray(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		if !IsZero(v.Index(i)) {
			return false
		}
	}
	return true
}

func isZeroStruct(v reflect.Value) bool {
	for i := 0; i < v.NumField(); i++ {
		if !IsZero(v.Field(i)) {
			return false
		}
	}
	return true
}
