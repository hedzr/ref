// Copyright © 2020 Hedzr Yeh.

// +build go1.13

// go1.14

package ref

import "reflect"

// IsZero reports whether v is the zero value for its type.
// It panics if the argument is invalid.
func IsZero(v reflect.Value) bool {
	// switch v.Kind() {
	// case reflect.Bool:
	// 	break
	// }
	return v.IsZero()
}

// IsNil reports whether its argument v is nil. The argument must be
// a chan, func, interface, map, pointer, or slice value; if it is
// not, IsNil panics. Note that IsNil is not always equivalent to a
// regular comparison with nil in Go. For example, if v was created
// by calling ValueOf with an uninitialized interface variable i,
// i==nil will be true but v.IsNil will panic as v will be the zero
// Value.
func IsNil(v reflect.Value) bool {
	// switch v.Kind() {
	// case reflect.Bool:
	// 	break
	// }
	return v.IsNil()
}
