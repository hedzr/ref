package ref

import (
	"fmt"
	"reflect"
	"strconv"
)

// Dump prints the structure and value of an object with pretty format
func Dump(obj interface{}, objDesc string, dumper func(level int, desc string, v reflect.Value)) {
	DumpEx(obj, objDesc, dumper, nil, nil)
}

// DumpEx prints the structure and value of an object with pretty format
func DumpEx(obj interface{}, objDesc string, dumper func(level int, desc string, v reflect.Value), preDumper, postDumper func(v reflect.Value)) {
	v := reflect.ValueOf(obj)
	ctx := ctx{
		seen:        make(map[comparison]bool),
		seenRecords: make(map[reflect.Value]bool),
		objDesc:     objDesc,
		dumper:      dumper,
	}
	if postDumper != nil {
		defer postDumper(v)
	}
	if preDumper != nil {
		preDumper(v)
	} else {
		if objDesc == "" {
			objDesc = "object"
		}
		vt := "<nil>"
		if obj != nil {
			vt = v.Type().String()
		}
		desc := fmt.Sprintf("Dumping %q (%v):", objDesc, vt)
		dumper(-1, desc, v)
	}
	dump(ctx, 0, v)
}

//func addValueLog(seen map[comparison]bool, v reflect.Value) (added bool) {
//	if !CanIsNil(v) || !IsNil(v) {
//		added = equal(v, k, seen)
//	}
//}

type ctx struct {
	parent      *ctx
	seen        map[comparison]bool
	seenRecords map[reflect.Value]bool
	objDesc     string
	dumper      func(level int, desc string, v reflect.Value)
}

// dump is a helper function.
//
// It's inspired from *gopl* chapter 12 - display(...)
//
// https://github.com/adonovan/gopl.io
func dump(c ctx, level int, v reflect.Value) {
	if v.CanAddr() {
		var z interface{}
		if v.CanInterface() {
			z = v.Interface()
		}
		var canIsNil = CanIsNil(v)
		var isNil = canIsNil && v.IsNil()
		if !canIsNil || !isNil {
			if z != nil {
				for k := range c.seenRecords {
					if equal(v, k, c.seen) {
						c.dumper(level, fmt.Sprintf("%s -> %v // <circular link detected, ignored>",
							c.objDesc, z), v)
						return
					}
				}
				c.seenRecords[v] = true
			}
		}
	}

	switch v.Kind() {
	case reflect.Invalid:
		c.dumper(level, fmt.Sprintf("%s = <invalid>", c.objDesc), v)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			cc := ctx{&c, c.seen, c.seenRecords, fmt.Sprintf("%s[%d]", c.objDesc, i), c.dumper}
			dump(cc, level+1, v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fieldPath := fmt.Sprintf("%s.%s", c.objDesc, v.Type().Field(i).Name)
			cc := ctx{&c, c.seen, c.seenRecords, fieldPath, c.dumper}
			dump(cc, level+1, v.Field(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			cc := ctx{&c, c.seen, c.seenRecords, fmt.Sprintf("%s[%s]", c.objDesc,
				formatAtom(key)), c.dumper}
			dump(cc, level+1, v.MapIndex(key))
		}
	case reflect.Ptr:
		if v.IsNil() {
			c.dumper(level, fmt.Sprintf("%s = nil", c.objDesc), v)
		} else {
			// cc:=ctx{&c,c.seen, fieldPath,c.dumper}
			cc := ctx{&c, c.seen, c.seenRecords, fmt.Sprintf("(*%s)", c.objDesc), c.dumper}
			dump(cc, level+1, v.Elem())
		}
	case reflect.Interface:
		if v.IsNil() {
			c.dumper(level, fmt.Sprintf("%s = nil", c.objDesc), v)
		} else {
			cc := ctx{&c, c.seen, c.seenRecords, c.objDesc + ".value", c.dumper}
			//fmt.Printf("%s.type = %s\n", c.objDesc, v.Elem().Type())
			c.dumper(level, fmt.Sprintf("%s.type = %s", c.objDesc, v.Elem().Type()), v)
			dump(cc, level+1, v.Elem())
		}
	default: // basic types, channels, functions
		c.dumper(level, fmt.Sprintf("%s = %s", c.objDesc, formatAtom(v)), v)
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
	// case reflect.Complex64, reflect.Complex128:
	// 	return strconv.FormatComplex(v.Complex(), 'f', -1, 128)
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
