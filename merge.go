package ref

import (
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// map merging

type (
	Merger struct {
		m                     Value
		ec                    *errors.WithCauses
		IgnoreUnexportedError bool
	}

	context struct {
		fromOrig, toOrig Value
		from, to         Value
	}
)

// NewMerger return a merging operator for your input map instance.
//
//     m := map[string]interface( "a": 1 }
//     var targetMap map[string]interface{}
//     NewMerger(m).MergeTo(&targetMap)
//
func NewMerger(inputMap interface{}) *Merger {
	mm := &Merger{
		m:                     ValueOf(inputMap),
		ec:                    errors.NewContainer(""),
		IgnoreUnexportedError: true,
	}
	//if mm.m.Kind() != reflect.Map {
	//	mm.ec.Attach(errors.New("inputMap MUST BE a map object or its ref.Value representation"))
	//} else {
	//	tk := mm.m.Type().Key()
	//	if tk.Kind() != reflect.String {
	//		mm.ec.Attach(errors.New("inputMap MUST BE a map[string]... object"))
	//	}
	//}
	return mm
}

//func MergeMap(from, to reflect.Value) {}

func (m *Merger) HasError() bool { return !m.ec.IsEmpty() }
func (m *Merger) Reset()         { m.ec = errors.NewContainer("") }

func (m *Merger) MergeTo(to interface{}) (err error) {
	if m.ec.IsEmpty() {
		m.ec.Attach(m.merge(m.m, ValueOf(to)))
	}
	return m.ec.Error()
}

func (m *Merger) merge(from, to Value) (err error) {
	err = m.impl(newContext(from, to))
	return
}

func (m *Merger) impl(c *context) (err error) {
	//if !c.from.Type().AssignableTo(c.to.Type()) {
	//	err = errors.New("cannot assign from %v to %v", c.from.Type(), c.to.Type())
	//	return
	//}

	//if c.to.Kind() == reflect.Interface {
	//	c.to = ValueOf(c.to.GetValue())
	//}

	fromKind := c.from.Kind()
	switch fromKind {
	// case reflect.Ptr: // never
	case reflect.Slice, reflect.Array:
		err = m.mergeSliceTo(c.from, c.to, c.to.Type(), nil)
		return
	case reflect.Struct:
		err = m.mergeStructTo(c.from, c.to, c.to.Type(), nil)
		return
	case reflect.Map:
		mapKeys := c.from.MapKeys()
		for _, key := range mapKeys {
			val := c.from.MapIndex(key)
			err = m.mergeValInto(c, key, val, c.to)
			if err != nil {
				return
			}
		}
		return
	default:
		if c.from.Type().AssignableTo(c.to.Type()) {
			//log.Debugf("        copying %v -> %v, simple set.", from.Type(), tot)
			defer m.deferRecoverFunc(&err, func(e interface{}) error {
				return errors.New("failed on copying %v -> %v (a), simple set. inner error: %v", c.from.Type(), c.to.Type(), e)
			})
			c.to.Set(c.from.Value)
			return
		} else if c.from.Kind() == c.to.Kind() {
			// log.Debugf("        copying %v -> %v, simple set.", from.Type(), to.Type())
			defer m.deferRecoverFunc(&err, func(e interface{}) error {
				return errors.New("failed on copying %v -> %v (a), simple set. inner error: %v", c.from.Type(), c.to.Type(), e)
			})
			c.to.Set(c.from.Value)
			return
		} else {
			var out reflect.Value
			if out, err = tryConvert(c.from.Value, c.to.Type()); err == nil {
				c.to.Set(out)
				return
			}
		}
	}
	err = errors.New("cannot merge from %v to %v (not impl)", c.from.Type(), c.to.Type())
	return
}

func (m *Merger) mergeValInto(c *context, key, val reflect.Value, to Value) (err error) {
	value := interfaceToRealType(Value{val})

	switch to.Kind() {
	case reflect.Map:
		err = m.mergeValIntoMap(c, key, value, to)
	case reflect.Slice, reflect.Array:
		err = m.mergeValIntoSlice(c, key, value, to)
	case reflect.Struct:
		err = m.mergeValIntoStruct(c, key, value, to)
	case reflect.Ptr:
		err = errors.New("cannot merge val into ptr %v", to.Type())
	default:
		err = errors.New("cannot merge val into %v,%v", to.Kind(), to.Type())
	}
	return
}

func (m *Merger) mergeValIntoSlice(c *context, key reflect.Value, val, toSlice Value) (err error) {
	err = errors.New("cannot merge val into slice %v", toSlice.Type())
	return
}

func (m *Merger) mergeValIntoStruct(c *context, key reflect.Value, value, toStruct Value) (err error) {
	var v1 = Value{key}
	var v1v = v1.GetValue()
	var v1key string
	if vk, ok := v1v.(string); ok {
		v1key = vk
	} else if vk, ok := v1v.(interface{ String() string }); ok {
		v1key = vk.String()
	} else {
		err = errors.New("expecting key is a stringer but it's %v", key.Type())
		return
	}

	if toFieldType, ok := toStruct.Type().FieldByName(v1key); ok {
		toField := toStruct.FieldByName(v1key)
		err = m.mergeValIntoStructField(c, key, value, toStruct, Value{toField}, toFieldType)
	} else {
		v1key = Captalize(v1key)
		if toFieldType, ok = toStruct.Type().FieldByName(v1key); ok {
			toField := toStruct.FieldByName(v1key)
			err = m.mergeValIntoStructField(c, key, value, toStruct, Value{toField}, toFieldType)
		}
		// err = errors.New("no field %q found in target struct", v1key)
	}
	return
}

func (m *Merger) mergeValIntoStructField(c *context, key reflect.Value, val, toStruct, toField Value, toFieldType reflect.StructField) (err error) {
	switch val.Kind() {
	case reflect.Map:
		err = m.mergeMapToStructField(key, val, toField, toFieldType)
	case reflect.Slice, reflect.Array:
		err = m.mergeSliceToStructField(key, val, toField, toFieldType)
	case reflect.Struct:
		err = m.mergeStructToStructField(key, val, toField, toFieldType)
	// case reflect.Ptr:
	default:
		if val.Type().AssignableTo(toFieldType.Type) {
			toField.Set(val.Value)
			log.Debugf("        > merged %v (%v) -> field %q (%v)", val.GetValue(), val.Type(), toFieldType.Name, toFieldType.Type)
			return
		} else {
			var out reflect.Value
			out, err = tryConvert(val.Value, toFieldType.Type)
			if err == nil {
				toField.Set(out)
				return
			}
			err = errors.New("cannot merge val into struct field %q, %v", toFieldType.Name, toFieldType.Type).Attach(err)
			return
		}
	}
	err = errors.New("cannot merge val into struct field %q, %v", toFieldType.Name, toFieldType.Type)
	return
}

func (m *Merger) mergeMapToStructField(key reflect.Value, val, toField Value, toFieldType reflect.StructField) (err error) {
	err = errors.New("cannot merge map into field %q (%v)", toFieldType.Name, toFieldType.Type)
	return
}

func (m *Merger) mergeSliceToStructField(key reflect.Value, val, toField Value, toFieldType reflect.StructField) (err error) {
	err = errors.New("cannot merge slice into field %q (%v)", toFieldType.Name, toFieldType.Type)
	return
}

func (m *Merger) mergeStructToStructField(key reflect.Value, val, toField Value, toFieldType reflect.StructField) (err error) {
	err = errors.New("cannot merge struct into field %q (%v)", toFieldType.Name, toFieldType.Type)
	return
}

func (m *Merger) mergeValIntoMap(c *context, key reflect.Value, val, tgtMap Value) (err error) {
	if !key.Type().AssignableTo(tgtMap.Type().Key()) {
		err = errors.New("cannot assign map[%v] to map[%v]", key.Type(), tgtMap.Type().Key())
		return
	}

	valKind := val.Kind()
	var v1 = Value{key}
	//var v2 = Value{val}
	//if valKind == reflect.Interface {
	//	v2 = ValueOf(v2.GetValue())
	//	valKind = v2.Kind()
	//}

	log.Debugf("    copying field '%v' (value=%v)...", v1.GetValue(), val.GetValue())

	if tgtMap.IsNil() {
		//log.Debugf("tgtMap's ptr = %v", tgtMap.Addr().Type())
		newMap := reflect.MakeMap(tgtMap.Type())
		tgtMap.Addr().Elem().Set(newMap)
	}

	switch valKind {
	case reflect.Bool:
		tgtMap.SetMapIndex(key, val.Value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		tgtMap.SetMapIndex(key, val.Value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		tgtMap.SetMapIndex(key, val.Value)
	case reflect.Float32, reflect.Float64:
		tgtMap.SetMapIndex(key, val.Value)
	case reflect.Complex64, reflect.Complex128:
		tgtMap.SetMapIndex(key, val.Value)
	case reflect.String:
		tgtMap.SetMapIndex(key, val.Value)
	case reflect.Slice, reflect.Array:
		err = m.mergeSliceInto(v1, val, tgtMap)
	case reflect.Map:
		err = m.mergeMapInto(v1, val, tgtMap)
	case reflect.Ptr:
		err = m.mergePtrInto(v1, val, tgtMap)
	default:
		panic(errors.New("copying into map[%v], unknown source type %v (kind=%v, value=%v)", v1.GetValue(), val.Type(), valKind, val.GetValue()))
	}
	return
}

func (m *Merger) mergeSliceInto(key, valSlice Value, tgtMap Value) (err error) {
	vt := Value{tgtMap.MapIndex(key.Value)}

	vtReal := ValueOf(vt.GetValue())
	valid := vtReal.IsValid()
	zero := vtReal.IsZero()

	if valid && !zero && vtReal.Kind() != reflect.Slice {
		err = errors.New("mergeSliceInto: target (map['%v']) is not a slice (type=%v, zero=%v, valid=%v).", key.GetValue(), vtReal.Type(), zero, valid)
		return
	}

	if !valid || zero {
		var l int
		if vt.IsValid() {
			l = vt.Len()
		}
		log.Debugf("        > target slice is empty or invalid, simple put. len=%v. vt=%v, key=%v (%v). %v|%v", l, vt.GetValue(), key.GetValue(), key.Type(), tgtMap.GetValue().(map[string]interface{})["e"], tgtMap.Type())
		tgtMap.SetMapIndex(key.Value, valSlice.Value)
		return
	}

	//err = m.mergeSliceIntoSlice(valSlice, vtReal, tgtMap, key)
	err = m.mergeSliceToSlice(valSlice, vtReal, func(val Value) Value {
		tgtMap.SetMapIndex(key.Value, val.Value)
		log.Debugf("        new slice: %v", tgtMap.MapIndex(key.Value))
		return interfaceToRealType(Value{tgtMap.MapIndex(key.Value)})
	})
	return
}

//func (m *Merger) mergeSliceIntoSlice(src Value, realTarget Value, tgtMap, key Value) (err error) {
//	// realTarget := ValueOf(tgt.GetValue())
//	//if tgt.Kind() != reflect.Slice {
//	//	log.Debugf("        realTarget.type=%v, realTarget=%v", realTarget.Type(), realTarget.GetValue())
//	//	tgt = Value{tgt.Convert(realTarget.Type())}
//	//}
//
//	//seen := make(map[comparison]bool)
//	//for si := 0; si < src.Len(); si++ {
//	//	sv := src.Index(si)
//	//	var found bool
//	//	for i := 0; i < realTarget.Len(); i++ {
//	//		v := realTarget.Index(i)
//	//		//log.Debugf("        - comparing '%v' and '%v'", Value{sv}.GetValue(), Value{v}.GetValue())
//	//		if equal(v, sv, seen) {
//	//			//log.Debugf("          matched")
//	//			found = true
//	//			break
//	//		}
//	//	}
//	//	if !found {
//	//		ns := reflect.Append(realTarget.Value, sv)
//	//		realTarget = Value{ns}
//	//		tgtMap.SetMapIndex(key.Value, realTarget.Value)
//	//		log.Debugf("        new slice: %v", tgtMap.MapIndex(key.Value))
//	//	}
//	//}
//
//	m.mergeSliceToSlice(src, realTarget, func(val Value) Value {
//		tgtMap.SetMapIndex(key.Value, realTarget.Value)
//		log.Debugf("        new slice: %v", tgtMap.MapIndex(key.Value))
//		return interfaceToRealType(tgtMap.MapIndex(key.Value))
//	})
//	return
//}

func (m *Merger) mergeMapInto(key, valMap Value, tgtMap Value) (err error) {
	vt := tgtMap.MapIndex(key.Value)
	if !vt.IsValid() || vt.IsNil() {
		tgtMap.SetMapIndex(key.Value, valMap.Value)
		return
	}

	err = m.mergeMapToMap(valMap, Value{vt})
	return
}

func (m *Merger) mergePtrInto(key, val Value, tgtMap Value) (err error) {
	to := valueFromMap(tgtMap, key)
	toType := val.Type()
	if !to.IsValid() || to.IsNil() {
		vv := val.IndirectValueRecursive()
		target := reflect.New(vv.Type())
		log.Debugf("        tmp target is: %v; %+v", val.Type(), Value{target}.GetValue())
		if err = DefaultCloner.Copy(vv, target); err == nil {
			tgtMap.SetMapIndex(key.Value, target.Elem())
		} else {
			log.Errorf("copying ptr to ptr not ok: %v", err)
		}
		return
	}

	if !to.IsZero() {
		toType = to.Type()
	}
	err = m.mergeObjIntoObj(val.IndirectValueRecursive(), to.IndirectValueRecursive(), toType, func(val Value) Value {
		tgtMap.SetMapIndex(key.Value, val.Value)
		return valueFromMap(tgtMap, key)
	})
	return
}

func (m *Merger) mergeObjIntoObj(from, to Value, toType reflect.Type, setTo func(val Value) Value) (err error) {
	if !from.IsValid() {
		err = errors.New("invalid source value: %v", from.Type())
		return
	}

	if from.IsZero() {
		SetZero(to.Value)
		return
	}

	fk := from.Kind()
	switch fk {
	case reflect.Struct:
		err = m.mergeStructTo(from, to, toType, setTo)
	case reflect.Slice: // impossible entry
	case reflect.Map: // impossible entry
	case reflect.Ptr: // impossible entry
	default:
		panic(errors.New("not implemented for source type: %v %v", fk, from.Type()))
	}
	return
}

func (m *Merger) mergeStructTo(from, to Value, toType reflect.Type, setTo func(val Value) Value) (err error) {
	if !to.IsValid() {
		newTargetType, newToOrig, newTo := m.indirectCreate(toType)
		log.Debugf("        . newTargetType=%v, newToOrig.type=%v, newTo.type=%v", newTargetType, newToOrig.Type(), newTo.Type())
		// no := reflect.New(toType)
		defer m.deferRecoverFunc(&err, func(e interface{}) error {
			return errors.New("mergeStructTo: failed in setting new elem (type=%v) into 'to': %v", toType, e)
		})
		if setTo != nil {
			to = setTo(newTo)
		} else {
			to.Addr().Elem().Set(newTo.Value)
		}
		return
	}

	tk := to.Kind()
	switch tk {
	case reflect.Struct:
		err = m.mergeStructToStruct(from, to, toType, setTo)
	case reflect.Slice: // TODO struct->slice
		panic(errors.New("mergeStructTo: not implemented for source type: %v %v", tk, from.Type()))
	case reflect.Map: // TODO struct->map
		panic(errors.New("mergeStructTo: not implemented for source type: %v %v", tk, from.Type()))
	case reflect.Ptr: // TODO struct->ptr
		panic(errors.New("mergeStructTo: not implemented for source type: %v %v", tk, from.Type()))
	default:
		// TODO struct->others primitive types
		panic(errors.New("mergeStructTo: not implemented for source type: %v %v", tk, from.Type()))
	}
	return
}

func (m *Merger) mergeStructToStruct(from, to Value, toType reflect.Type, setTo func(val Value) Value) (err error) {
	// log.Debugf("        > struct %v (%v) -> struct %v (%v) ..", from.Type(), from.Type().PkgPath(), toType, toType.PkgPath())
	if strings.HasSuffix(to.Type().String(), "time.Time") {
		var tm []byte
		if strings.HasSuffix(from.Type().String(), "time.Time") {
			// tm = from.GetValue().(time.Time).UnixNano()
			tm, err = from.GetValue().(time.Time).MarshalBinary()
		}
		if err == nil {
			up := unsafe.Pointer(to.Addr().Pointer())
			err = ((*time.Time)(up)).UnmarshalBinary(tm)
		}
		return
	}

	fieldsCount := from.Type().NumField()
	for i := 0; i < fieldsCount; i++ {
		field := from.Type().Field(i)
		if !isExportableField(field) {
			if !m.IgnoreUnexportedError {
				err = errors.New("target cannot be set (unexported field): field %q (src-value: %v)", field.Name, from.GetValue())
				return
			}
			continue
		}
		if field.Anonymous {
			fieldValue := from.Field(i)
			if err = m.mergeStructTo(Value{fieldValue}, to, toType, setTo); err != nil {
				// err = errors.New("nested structure on field %q", field.Name)
				return
			}
			continue
		}

		var (
			tot    reflect.Type
			toName = field.Name
		)
		if ttf, ok := to.Type().FieldByName(toName); ok {
			tot = ttf.Type
		} else {
			// field -> func, ...
			log.Debugf("        ignored field merging: %q %v -> %v", field.Name, field.Type, "unknown target field")
			continue
		}

		if err = m.mergeFieldToField(Value{from.Field(i)}, Value{to.FieldByName(toName)}, field, tot, setTo); err != nil {
			return
		}
	}
	return
}

func (m *Merger) mergeFieldToField(fromV, toV Value, srcField reflect.StructField, tot reflect.Type, setTo func(val Value) Value) (err error) {
	if !fromV.IsValid() {
		err = errors.New("invalid source value: %v", fromV.Type())
		return
	}
	toType := tot
	log.Debugf("        .. src field %q %v -> tot: %v | toV: %v %v (valid: %v)", srcField.Name, srcField.Type, tot, toV.Kind(), toV.Type(), toV.IsValid())
	//if srcField.Name == "Birthday" {
	//	log.Debug()
	//}
	if !toV.IsValid() || (toV.Kind() == reflect.Ptr && toV.IsZero()) {
		log.Debugf("        .. src field %q %v: new target %v", srcField.Name, srcField.Type, toType.Elem())
		if toV.Kind() == reflect.Ptr {
			no := reflect.New(toType.Elem())
			toV.Addr().Elem().Set(no)
			toType = toV.Elem().Type()
			log.Debugf("        .. toType %v", toType)
		} else {
			err = errors.New("invalid target value: %v", toV.Type())
			return
		}
	}
	if fromV.IsZero() {
		if setTo != nil {
			setTo(fromV)
		} else {
			toV.IndirectValue().SetZero()
		}
		return
	}

	from, to := fromV.IndirectValueRecursive(), toV.IndirectValueRecursive()
	fk, tk := from.Kind(), toType.Kind()
	switch fk {
	case reflect.Struct:
		err = m.mergeStructTo(from, to, toType, setTo)
	case reflect.Slice, reflect.Array:
		err = m.mergeSliceTo(from, to, toType, setTo)
	case reflect.Map:
		err = m.mergeMapTo(from, to, toType, setTo)
	case reflect.Ptr:
		panic(errors.New("not implemented for source type: %v %v", tk, from.Type()))
	default:
		if from.Type().AssignableTo(toType) {
			//log.Debugf("        copying %v -> %v, simple set.", from.Type(), tot)
			defer m.deferRecoverFunc(&err, func(e interface{}) error {
				return errors.New("failed on copying field %q (%v) %v -> %v (a), simple set. inner error: %v", srcField.Name, srcField.Type, from.Type(), toType, e)
			})
			if setTo != nil {
				setTo(from)
			} else {
				to.Set(from.Value)
			}
		} else if fk == tk {
			// log.Debugf("        copying %v -> %v, simple set.", from.Type(), to.Type())
			defer m.deferRecoverFunc(&err, func(e interface{}) error {
				return errors.New("failed on copying field %q (%v) %v -> %v (b), simple set. inner error: %v", srcField.Name, srcField.Type, from.Type(), toType, e)
			})
			if setTo != nil {
				setTo(from)
			} else {
				to.Set(from.Value)
			}
		} else {
			var out reflect.Value
			if out, err = tryConvert(from.Value, toType); err == nil {
				to.Set(out)
			} else {
				log.Debugf("        copying field %q (%v) %v -> %v (tk=%v), simple set.", srcField.Name, srcField.Type, from.Type(), toType, tk)
				panic(errors.New("not implemented for source type: %v %v", fk, from.Type()))
			}
		}
	}
	return
}

func (m *Merger) mergeMapTo(from, to Value, tot reflect.Type, setTo func(val Value) Value) (err error) {
	switch tot.Kind() {
	case reflect.Map:
		err = m.mergeMapToMap(from, to)
		return
	case reflect.Ptr:
	case reflect.Slice, reflect.Array:
	case reflect.Struct:
		err = m.mergeMapToStruct(from, to)
		return
	default:
	}
	panic(errors.New("not implemented for source type: %v (tot: %v)", from.Type(), tot))
	return
}

func (m *Merger) mergeMapToMap(from, to Value) (err error) {
	to = interfaceToRealType(to)
	err = m.impl(newContext(from, to))
	return
}

func (m *Merger) mergeMapToStruct(from, to Value) (err error) {
	to = interfaceToRealType(to)
	err = m.impl(newContext(from, to))
	return
}

func (m *Merger) mergeSliceTo(from, to Value, tot reflect.Type, setTo func(val Value) Value) (err error) {
	switch tot.Kind() {
	case reflect.Struct:
	case reflect.Slice, reflect.Array:
		err = m.mergeSliceToSlice(from, to, setTo)
		return
	case reflect.Map:
	case reflect.Ptr:
	default:
	}
	panic(errors.New("not implemented for source type: %v (tot: %v)", from.Type(), tot))
	return
}

func (m *Merger) mergeSliceToSlice(from, to Value, setTo func(val Value) Value) (err error) {
	seen := make(map[comparison]bool)
	for si := 0; si < from.Len(); si++ {
		sv := from.Index(si)
		var found bool
		for i := 0; i < to.Len(); i++ {
			v := to.Index(i)
			//log.Debugf("        - comparing '%v' and '%v'", Value{sv}.GetValue(), Value{v}.GetValue())
			if equal(v, sv, seen) {
				//log.Debugf("          matched")
				found = true
				break
			}
		}
		if !found {
			ns := reflect.Append(to.Value, sv)
			if setTo != nil {
				to = Value{ns}
				to = setTo(to)
			} else {
				to.Set(ns)
			}
			//tgtMap.SetMapIndex(key.Value, realTarget.Value)
			//log.Debugf("        new slice: %v", tgtMap.MapIndex(key.Value))
		}
	}
	return
}

func (m *Merger) deferRecoverFunc(err *error, buildError func(e interface{}) error) func() {
	return func() {
		if e := recover(); e != nil {
			if e2, ok := e.(error); ok {
				*err = e2
			} else {
				*err = buildError(e)
				// *err = errors.New("%v: failed in setting new elem (type=%v) into 'to': %v", title, toType, e)
				log.Errorf("recovering: %v", *err)
			}
		}
	}
}

func (m *Merger) indirectCreate(fromType reflect.Type) (newTargetType reflect.Type, parent, newTo Value) {
	newTargetType = fromType
	log.Debugf("    .. indirectCreate %v ...", fromType)

	if newTargetType.Kind() != reflect.Ptr {
		newTo = ValueOf(reflect.New(newTargetType))
		parent = newTo
		newTo = newTo.IndirectValueRecursive()
		return
	}

	var tp Value
	var tt = fromType.Elem()
	newTargetType, tp, newTo = m.indirectCreate(tt)
	parent = tp
	if tp.CanAddr() {
		parent = ValueOf(reflect.New(fromType))
		parent.Set(tp.Addr())
	}
	return
}

func interfaceToRealType(v Value) Value {
	if v.Kind() == reflect.Interface {
		v = ValueOf(v.GetValue())
	}
	return v
}

func valueFromMap(m, k Value) Value {
	v := Value{m.MapIndex(k.Value)}
	return interfaceToRealType(v)
}

func newContext(from, to Value) *context {
	return &context{
		fromOrig: from,
		toOrig:   to,
		from:     from.IndirectValueRecursive(),
		to:       to.IndirectValueRecursive(),
	}
}
