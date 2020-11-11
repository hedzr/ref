package ref

import (
	"reflect"
	"unsafe"
)

type Field struct {
	Type  reflect.StructField
	Value reflect.Value
}

func (field Field) GetUnexportedField() interface{} {
	return reflect.NewAt(field.Value.Type(), unsafe.Pointer(field.Value.UnsafeAddr())).Elem().Interface()
}

func (field Field) SetUnexportedField(value interface{}) {
	reflect.NewAt(field.Value.Type(), unsafe.Pointer(field.Value.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

func GetUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func SetUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
