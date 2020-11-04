package ref

import (
	"fmt"
	"reflect"
	"strconv"
)

// Dump prints the structure and value of an object
func Dump(obj interface{}, dumper func(level int, ownerPath, desc string, v reflect.Value)) {
	v := reflect.ValueOf(obj)
	dumper(0, "", "Dump of object:", v)
	dump(0, "", v, dumper)
}

// DumpEx prints the structure and value of an object
func DumpEx(obj interface{}, dumper func(level int, ownerPath, desc string, v reflect.Value), preDumper, postDumper func(v reflect.Value)) {
	v := reflect.ValueOf(obj)
	defer postDumper(v)
	preDumper(v)
	dump(0, "", v, dumper)
}

// dump is a helper function.
//
// It's inspired from *gopl* chapter 12 - display(...)
func dump(level int, path string, v reflect.Value, dumper func(level int, ownerPath, desc string, v reflect.Value)) {
	switch v.Kind() {
	case reflect.Invalid:
		dumper(level, path, fmt.Sprintf("%s = invalid\n", path), v)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			dump(level+1, fmt.Sprintf("%s[%d]", path, i), v.Index(i), dumper)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fieldPath := fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name)
			dump(level+1, fieldPath, v.Field(i), dumper)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			dump(level+1, fmt.Sprintf("%s[%s]", path,
				formatAtom(key)), v.MapIndex(key), dumper)
		}
	case reflect.Ptr:
		if v.IsNil() {
			dumper(level, path, fmt.Sprintf("%s = nil\n", path), v)
		} else {
			dump(level+1, fmt.Sprintf("(*%s)", path), v.Elem(), dumper)
		}
	case reflect.Interface:
		if v.IsNil() {
			dumper(level, path, fmt.Sprintf("%s = nil\n", path), v)
		} else {
			fmt.Printf("%s.type = %s\n", path, v.Elem().Type())
			dump(level+1, path+".value", v.Elem(), dumper)
		}
	default: // basic types, channels, funcs
		dumper(level, path, fmt.Sprintf("%s = %s\n", path, formatAtom(v)), v)
	}
}

// Any formats any value as a string.
func Any(value interface{}) string {
	return formatAtom(reflect.ValueOf(value))
}

// formatAtom formats a value without inspecting its internal structure.
func formatAtom(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Invalid:
		return "invalid"
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Complex64, reflect.Complex128:
		return strconv.FormatComplex(v.Complex(), 'f', -1, 128)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return strconv.Quote(v.String())
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Slice, reflect.Map:
		return v.Type().String() + " 0x" +
			strconv.FormatUint(uint64(v.Pointer()), 16)
	default: // reflect.Array, reflect.Struct, reflect.Interface
		return v.Type().String() + " value"
	}
}
