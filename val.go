package ref

import (
	"gopkg.in/hedzr/errors.v2"
	"math"
	"reflect"
	"unsafe"
)

// Value is a struct wrapped on reflect.Value
type Value struct {
	reflect.Value
}

// ValueOf is a copy of reflect.ValueOf(i) but wrapped as ref.Value
//
// both the original object and its reflect.Value obj can be passed into ValueOf.
//
//     import "github.com/hedzr/assert"
//     import "github.com/hedzr/ref"
//     import "reflect"
//     a := &User{Name:"me"}
//     va := reflect.ValueOf(a)
//     vv1 := ref.ValueOf(a)
//     vv2 := ref.ValueOf(va)
//     assert.Equal(t, vv1, vv2) // vv1 and vv2 should be always equivalent.
//
func ValueOf(obj interface{}) Value {
	if v, ok := obj.(reflect.Value); ok {
		return Value{v}
	}
	if v, ok := obj.(Value); ok {
		return v
	}
	return Value{reflect.ValueOf(obj)}
}

func isKindInt(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func isKindUint(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isKindInteger(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isKindFloat(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func isKindComplex(k reflect.Kind) bool {
	switch k {
	case reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

func isStruct(obj interface{}) bool  { return reflect.TypeOf(obj).Kind() == reflect.Struct }
func isPointer(obj interface{}) bool { return reflect.TypeOf(obj).Kind() == reflect.Ptr }

func (v Value) IsKindInt(k reflect.Kind) bool     { return isKindInt(k) }
func (v Value) IsKindUint(k reflect.Kind) bool    { return isKindUint(k) }
func (v Value) IsKindInteger(k reflect.Kind) bool { return isKindInteger(k) }
func (v Value) IsKindFloat(k reflect.Kind) bool   { return isKindFloat(k) }
func (v Value) IsKindComplex(k reflect.Kind) bool { return isKindComplex(k) }
func (v Value) IsKindStruct(k reflect.Kind) bool  { return k == reflect.Struct }
func (v Value) IsKindPointer(k reflect.Kind) bool { return k == reflect.Ptr }
func (v Value) IsString() bool                    { return v.Kind() == reflect.String }
func (v Value) IsInt() bool                       { return v.IsKindInt(v.Kind()) }
func (v Value) IsFloat() bool                     { return v.IsKindFloat(v.Kind()) }
func (v Value) IsComplex() bool                   { return v.IsKindComplex(v.Kind()) }
func (v Value) IsStruct() bool                    { return v.Kind() == reflect.Struct }
func (v Value) IsPointer() bool                   { return v.Kind() == reflect.Ptr }
func (v Value) IndirectType() Type                { return Type{v.Type()}.IndirectType() }
func (v Value) IndirectTypeRecursive() Type       { return Type{v.Type()}.IndirectTypeRecursive() }

func (v Value) IndirectValue() Value {
	if v.Kind() == reflect.Ptr {
		return Value{v.Elem()}
	}
	return v
}

func IndirectValue(reflectValue reflect.Value) reflect.Value {
	if reflectValue.Kind() == reflect.Ptr {
		return reflectValue.Elem()
	}
	return reflectValue
}

func (v Value) IndirectValueRecursive() Value {
	x := v
	for x.Kind() == reflect.Ptr {
		x = Value{x.Elem()}
	}
	return x
}

func (v Value) PtrToIndirectValueRecursive() Value {
	vv, last := v, v
	for vv.Kind() == reflect.Ptr {
		last = vv
		vv = Value{vv.Elem()}
	}
	return last
}

func IndirectValueRecursive(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

// IsZero reports whether v is the zero value for its type.
// It panics if the argument is invalid.
func (v Value) IsZero() bool {
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
		return v.IsNil()
	case reflect.UnsafePointer:
		return v.IsNil()
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

func (v Value) IsNil() bool {
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.UnsafePointer:
		ptr := (unsafe.Pointer)(v.Pointer())
		return ptr == nil
	case reflect.Map:
		return v.Value.Len() == 0
	//case ...:
	//	//if v.flag&flagMethod != 0 {
	//	//	return false
	//	//}
	//	ptr := (unsafe.Pointer)(v.Pointer())
	//	//ptr := v.ptr
	//	//if v.flag&flagIndir != 0 {
	//	//	ptr = *(*unsafe.Pointer)(ptr)
	//	//}
	//	return ptr == nil
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Interface:
		return v.Interface() == nil
		//case reflect.Interface, reflect.Slice:
		//	// Both interface and slice are nil if first word is 0.
		//	// Both are always bigger than a word; assume flagIndir.
		//	return *(*unsafe.Pointer)(v.ptr) == nil
	}
	return false
}

func (v Value) SetZero() {
	if CanIsZero(v.Value) {
		v.Set(reflect.Zero(v.Type()))
	}
}

func (v Value) SetNil() {
	//k := v.Kind()
	//switch k {
	//case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer:
	//	SetNil(v.Value)
	//case reflect.Slice:
	//	SetNil(v.Value)
	//case reflect.Interface:
	//	SetNil(v.Value)
	//}

	// SetNil(v.Value)

	if CanIsNil(v.Value) {
		// to.Set(reflect.New(to.Type().Elem()))
		z := reflect.Zero(v.Type())
		v.Set(z)
	}
}

func (v Value) GetValue() (val interface{}) {
	if v.IsValid() {
		if v.CanInterface() {
			val = v.Interface()
		} else {
			k := v.Kind()
			switch k {
			case reflect.Bool:
				val = v.Bool()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val = v.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				val = v.Uint()
			case reflect.Float32, reflect.Float64:
				val = v.Float()
			case reflect.Complex64, reflect.Complex128:
				val = v.Complex()
			default:
				panic(errors.New("unknown kind %v", k))
			}
		}
	}
	return
}

// TryConvert calls reflect.Convert safely, without panic threw.
func (v Value) TryConvert(t reflect.Type) (out reflect.Value, err error) {
	return TryConvert(v.Value, t)
}

func reflectValueRecursive(obj interface{}) reflect.Value {
	var reflectValue reflect.Value = reflect.ValueOf(obj)
	return IndirectValueRecursive(reflectValue)
}

func reflectTypeRecursive(obj interface{}) reflect.Type {
	var reflectType reflect.Type = reflect.TypeOf(obj)
	return IndirectTypeRecursive(reflectType)
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}
