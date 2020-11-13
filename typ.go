package ref

import "reflect"

// TypeOf returns the reflection Type that represents the dynamic type of i.
// If i is a nil interface value, TypeOf returns nil.
func TypeOf(obj interface{}) Type {
	if v, ok := obj.(reflect.Type); ok {
		return Type{v}
	}
	if v, ok := obj.(Type); ok {
		return v
	}
	return Type{reflect.TypeOf(obj)}
}

// Type is a struct wrapped on reflect.Type
type Type struct {
	reflect.Type
}

func (t Type) IndirectType() Type {
	if t.Kind() == reflect.Ptr {
		return Type{t.Elem()}
	}
	return t
}

func IndirectType(reflectType reflect.Type) reflect.Type {
	if reflectType.Kind() == reflect.Ptr {
		return reflectType.Elem()
	}
	return reflectType
}

func (t Type) IndirectTypeRecursive() Type {
	x, xk := t, t.Kind()
	for xk == reflect.Ptr || xk == reflect.Slice {
		x = Type{x.Elem()}
		xk = x.Kind()
	}
	return x
}

func IndirectTypeRecursive(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}
