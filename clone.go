package ref

import (
	"bytes"
	"encoding/gob"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	"strings"
	"unicode"
)

type (
	// Cloner provides a copier interface for the deep clone algor.
	Cloner interface {
		Copy(fromVar, toVar interface{}, Opts ...Opt) (err error)
	}

	// Opt is functional option functor for Cloner
	Opt func(c Cloner)

	// CloneOpt is functional option functor for Clone()
	CloneOpt func(c *cloner)

	// Cloneable interface represents a cloneable object
	Cloneable interface {
		// Clone will always return a new cloned object instance on 'this' object
		Clone() interface{}
	}

	//
	NameMappingRule func(fromName string) (toName string, mapped bool)
)

var (
	DefaultNameMappingRule NameMappingRule = func(fromName string) (toName string, mapped bool) {
		return
	}

	DefaultCloner = buildDefaultCloner()

	LazyGobCopier = lazyCopier{}
)

// Clone makes a deep clone of 'from'.
//
// Note that toVar should be a pointer to a struct while you're cloning on a struct, for example:
//
//     var u1 = User{Name:"aa"}
//     var u2 = new(User)
//     fmt.Println(ref.Clone(u1, u2))
//
// Clone will make the cloned object from fromVar with Cloneable interface. It's talking about
// the better way to clone an object is implement the Cloneable interface.
//
// For the normally object cloning, Clone uses Golang reflect feature. So we can only clone the
// exported fields of a struct in this case. The unexported fields will be ignored by
// default ( see also cloner.IgnoreUnexportedError ) .
//
func Clone(fromVar, toVar interface{}, opts ...CloneOpt) interface{} {
	c := buildDefaultCloner()
	for _, opt := range opts {
		opt(&c)
	}

	if fromVar == nil {
		if toVar != nil {
			ValueOf(toVar).IndirectValue().SetNil() // set the target object to zero or nil
		}
		return toVar
	}
	if toVar == nil {
		return toVar
	}

	if err := c.Copy(fromVar, toVar); err != nil {
		log.Warnf("Clone not ok: %v", err)
	}
	return toVar
}

// WithIgnoredFieldNames appends the ignored name list to the Cloner operator.
func WithIgnoredFieldNames(names ...string) CloneOpt {
	return func(c *cloner) {
		c.ignoredNames = append(c.ignoredNames, names...)
	}
}

func WithNameMappings(m map[string]string) CloneOpt {
	return func(c *cloner) {
		c.nameMappingMap = m
	}
}

func WithNameMappingsRule(rule NameMappingRule) CloneOpt {
	return func(c *cloner) {
		c.nameMappingRule = rule
	}
}

func WithIgnoreUnexportedError(ignored bool) CloneOpt {
	return func(c *cloner) {
		c.IgnoreUnexportedError = ignored
	}
}

func WithZeroIfEqualsFrom(b bool) CloneOpt {
	return func(c *cloner) {
		c.ZeroIfEqualsFrom = b
	}
}

func WithKeepIfFromIsNilOrZero(keepIfFromIsNil, keepIfFromIsZero bool) CloneOpt {
	return func(c *cloner) {
		c.KeepIfFromIsNil = keepIfFromIsNil
		c.KeepIfFromIsZero = keepIfFromIsZero
	}
}

func WithEachFieldAlways(b bool) CloneOpt {
	return func(c *cloner) {
		c.EachFieldAlways = b
	}
}

func buildDefaultCloner() cloner {
	return cloner{
		ignoredNames:          nil,
		nameMappingMap:        make(map[string]string),
		nameMappingRule:       DefaultNameMappingRule,
		IgnoreUnexportedError: true,
		ZeroIfEqualsFrom:      false,
		KeepIfFromIsNil:       false,
		KeepIfFromIsZero:      false,
		EachFieldAlways:       false,
	}
}

// lazyCopier is a Cloner by using encoding/gob, it's just available for those exported fields.
type lazyCopier struct {
}

func (c lazyCopier) Copy(fromVar, toVar interface{}, Opts ...Opt) (err error) {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	if err = enc.Encode(fromVar); err == nil {
		err = dec.Decode(toVar)
	}
	return
}

// cloner is a standard Cloner by using golang reflect library.
type cloner struct {
	ignoredNames          []string
	nameMappingMap        map[string]string
	nameMappingRule       NameMappingRule
	IgnoreUnexportedError bool
	ZeroIfEqualsFrom      bool // 源和目标字段值相同时，目标字段被清除为未初始化的零值
	KeepIfFromIsNil       bool // 源字段值为nil指针时，目标字段的值保持不变
	KeepIfFromIsZero      bool // 源字段值为未初始化的零值时，目标字段的值保持不变 // 此条尚未实现
	EachFieldAlways       bool
}

func (c cloner) Copy(fromVar, toVar interface{}, Opts ...Opt) (err error) {
	if !hasAnyValidTypes(fromVar, reflect.Struct, reflect.Ptr, reflect.Map) {
		err = errors.New("fromVar should be a valid struct object or pointer object")
		return
	}
	if !hasAnyValidTypes(toVar, reflect.Struct, reflect.Ptr, reflect.Map) {
		err = errors.New("toVar should be a valid struct object or pointer object")
		return
	}

	fv, tv := ValueOf(fromVar), ValueOf(toVar)
	ft, tt := fv.Type(), tv.Type()
	// from := reflectValueRecursive(fromVar)
	// to := reflectValueRecursive(toVar)
	from, to := fv.IndirectValueRecursive(), tv.IndirectValueRecursive()

	// modelType := reflect.TypeOf((*Cloneable)(nil)).Elem()
	// if objType.Implements(modelType) {
	//	//
	// }
	if z, ok := fromVar.(Cloneable); ok {
		err = c.copyCloneableObject(z, tv, fromVar, toVar)
		return
	}

	if !to.CanAddr() {
		return errors.New("clone to value is unaddressable")
	}

	if !from.IsValid() {
		return errors.New("clone to value is unaddressable")
	}

	if from.IsNil() || from.IsZero() {
		// SetNil(to)
		// SetZero(to)
		to.Set(reflect.Zero(to.Type()))
		return
	}

	fk := from.Kind()
	if fk == reflect.Struct {
		err = c.copyStructTo(from, to, ft, tt, fv, tv)
	} else if fk == reflect.Slice || fk == reflect.Array {
		err = c.copySliceTo(from, to, fk, to.Kind(), ft, tt, fv, tv)
	} else if fk == reflect.Map {
		err = c.copyMapTo(from, to, fk, to.Kind(), ft, tt, fv, tv)
	} else {
		sfk, stk := ft.Kind(), tt.Kind()
		log.Warnf("not implemented: %v (%v) -> %v (%v)", ft, sfk, tt, stk)
		// []ref.Employee (slice) -> ref.User (struct)
	}
	return
}

func (c cloner) copyMapTo(from, to Value, fk, tk reflect.Kind, oft, ott reflect.Type, ofv, otv Value) (err error) {
	fromType, toType := from.Type(), to.Type()
	keyType := fromType.Key()
	valType := fromType.Elem()
	log.Debug("source map: %v. k=%v, v=%v", fromType, keyType, valType)

	switch toType.Kind() {
	case reflect.Map:
		err = c.copyMapToMap(from, to, fromType, toType, keyType, valType)
	case reflect.Struct:
		err = c.copyMapToStruct(from, to, fromType, toType, keyType, valType)
	default:
		err = errors.New("copying from %v to %v: NOT IMPLEMENT", oft, ott)
	}

	// err = errors.New("not implement")
	return
}

func (c cloner) copyMapToStruct(from, to Value, fromType, toType, keyType, valType reflect.Type) (err error) {
	if to.IsNil() {
		nsi := reflect.New(toType)
		to.Addr().Elem().Set(nsi)
	}

	var ec = errors.NewContainer("copying")
	var stringType = reflect.TypeOf("")

	for _, key := range from.MapKeys() {
		value := from.MapIndex(key)
		if value.CanInterface() {
			i := value.Interface()
			//log.Debugf("i: %v / %v. type: %v %v", i, ValueOf(i).Kind(), value.Type(), value.Kind())
			value = reflect.ValueOf(i)
		}

		k := key.Convert(stringType)
		name := k.String()
		structFieldValue := to.FieldByName(name)

		// TODO more target types. for example: map, slice, ptr to struct\map\slice, ...
		if value.Kind() == reflect.Map && structFieldValue.Kind() == reflect.Struct {
			err = c.copyMapToStruct(Value{value}, Value{structFieldValue}, value.Type(), structFieldValue.Type(), value.Type(), structFieldValue.Type())
			return
		}

		var ecf = errors.NewContainer("might be invalid field")

	adaptedNewFieldName:
		if !structFieldValue.IsValid() {
			if unicode.IsLower([]rune(name)[0]) || strings.Contains(name, "-") {
				name = strings.Title(name)
				structFieldValue = to.FieldByName(name)
				log.Debugf("    trying for field name %q", name)
				ecf = errors.NewContainer("might be invalid field")
				goto adaptedNewFieldName
			}
			ecf.Attach(errors.New("no such field: %s in obj", name))
			ec.Attach(ecf.Error())
			continue
		}

		// If obj field value is not settable an error is thrown
		if !structFieldValue.CanSet() {
			if unicode.IsLower([]rune(name)[0]) || strings.Contains(name, "-") {
				name = strings.Title(name)
				structFieldValue = to.FieldByName(name)
				log.Debugf("    trying field name %q", name)
				ecf = errors.NewContainer("field")
				goto adaptedNewFieldName
			}
			ecf.Attach(errors.New("cannot set %s field value", name))
			ec.Attach(ecf.Error())
			continue
		}

		structFieldType := structFieldValue.Type()
		//val := value.Convert(structFieldType)

		if !value.Type().AssignableTo(structFieldType) {
			value, err = c.tryConvert(value, structFieldType)
			if err != nil {
				ec.Attach(err)
				continue
			}
			//ecf.Attach(errors.New("provided value type '%v' didn't match obj field type '%v'", value.Type(), structFieldType))
			//ec.Attach(ecf.Error())
			//continue
		}

		structFieldValue.Set(value)
		ec.Attach(ecf.Error())
	}

	err = ec.Error()
	return
}

func (c cloner) copyMapToMap(from, to Value, fromType, toType, keyType, valType reflect.Type) (err error) {
	// 'to' mustBeMap and 'ptrToTo' mustBePtrToMap
	// log.Debugf("ptrToTo: %v, to: %v, toType: %v", ptrToTo.Type(), to.Type(), toType)
	if to.IsNil() { // map is nil?
		// if an empty map found, create a new one
		nmi := reflect.MakeMap(toType)
		to.Addr().Elem().Set(nmi)
	}

	for _, key := range from.MapKeys() {
		toKeyType := toType.Key()
		toValType := toType.Elem()
		//if !keyType.AssignableTo(toKeyType) || !valType.AssignableTo(toValType) {
		//	continue
		//}
		if key.CanInterface() {
			log.Debugf("  - %v:%v = %v", toKeyType, toValType, key.Interface())
		}
		k := key.Convert(toKeyType)
		v := from.MapIndex(key)
		log.Debugf("    k=%v, v=%v", k, v)
		to.SetMapIndex(k, v.Convert(toValType))
	}
	return
}

func (c cloner) copySliceTo(from, to Value, fk, tk reflect.Kind, oft, ott reflect.Type, ofv, otv Value) (err error) {
	err = errors.New("not implement")
	return
}

func (c cloner) copyStructTo(from, to Value, oft, ott reflect.Type, ofv, otv Value) (err error) {
	fromType := from.Type()

	if to.Kind() == reflect.Slice {
		// note that array have not been target yet

		baseType := to.Type().Elem()
		newTargetType, newToOrig, newTo := c.indirectCreate(baseType)
		log.Debugf(" copying struct to slice[0]: srcV=%v, tgtV=%v (baseType=%v, newTo.Type=%v, newToOrig.Type=%v)", fromType, newTargetType, baseType, newTo.Type(), newToOrig.Type())
		if err = c.copyStructTo(from, newTo, oft, newTargetType, ofv, newToOrig); err == nil {
			v := newToOrig.Value
			if baseType == newTargetType {
				v = newTo.Value
			}
			if to.Len() == 0 {
				to.Set(reflect.Append(to.Value, v))
			} else {
				to.Index(0).Set(v)
			}
		}
		return
	}

	var h held
	var toKind = to.Kind()
	if toKind == reflect.Map {
		//baseType := to.Type().Elem()
		//newTargetType, newToOrig, newTo := c.indirectCreate(baseType)
		//log.Debugf(" copying struct to slice[0]: srcV=%v, tgtV=%v (baseType=%v, newTo.Type=%v, newToOrig.Type=%v)", fromType, newTargetType, baseType, newTo.Type(), newToOrig.Type())
		h = newHeldMap(to.Value)
	} else if toKind == reflect.Struct {
		h = newHeldStruct(to.Value)
	} else {
		err = errors.New("cannot copy struct (%v) to %v", oft, ott.Elem())
		return
	}

	fieldsCount := fromType.NumField()
	for i := 0; i < fieldsCount; i++ {
		field := fromType.Field(i)
		if c.shouldBeIgnored(field.Name) {
			continue
		}
		if !isExportableField(field) {
			toName := c.targetName(field.Name)
			vov := from.Field(i)
			tof := to.FieldByName(toName)
			if tof.CanSet() {
				tof.Set(vov)
			} else if !c.IgnoreUnexportedError {
				err = errors.New("target cannot be set: field %q (value: %v)", field.Name, vov.Interface())
				return
			}
			continue
		}
		if field.Anonymous {
			fieldValue := from.Field(i)
			if err = c.copyStructTo(Value{fieldValue}, to, oft, ott, ofv, otv); err != nil {
				// err = errors.New("nested structure on field %q", field.Name)
				return
			}
			continue
		}

		var (
			tot    reflect.Type
			vov    = from.Field(i)
			toName = c.targetName(field.Name)
			tof    = h.Get(field, toName)
			// tof := to.FieldByName(toName)
		)
		if toKind == reflect.Struct {
			if ttf, ok := to.Type().FieldByName(toName); ok {
				tot = ttf.Type
				// log.Debugf("  | ttf: %v | tt: %v %v | tof.IsValid: %v } to: %v", ttf, tt.Kind(), tt, tof.IsValid(), to.Type())
			} else {
				for z := otv.Value; z.IsValid(); {
					tm := z.MethodByName(toName)
					if tm.IsValid() {
						tof, tot = tm, tm.Type()
						break
					}
					if z.Kind() != reflect.Ptr {
						break
					}
					z = z.Elem()
				}
				if tot == nil || tot.Kind() == reflect.Invalid {
					if !tof.IsValid() {
						continue // the target field not exists, ignore it and go to next field
					}

					log.Debugf("  | tof.IsValid: %v", tof.IsValid())
					tot = tof.Type()
				}
			}
		}
		h.SetTargetField(tof)

		// fk, tk := from.Kind(), to.Kind()
		if err = c.copyFieldToField(vov, h, field.Type, tot, field.Name, toName); err != nil {
			return
		}
	}

	methodCount := fromType.NumMethod()
	for i := 0; i < methodCount; i++ {
		method := fromType.Method(i)
		if c.shouldBeIgnored(method.Name) {
			continue
		}
		if (method.Type.NumIn() != 0 && method.Type.NumIn() != 1) || method.Type.NumOut() != 1 {
			// log.Debugf("  -> func %q : %v. in: %v, out: %v", method.Name, method.Type, method.Type.NumIn(), method.Type.NumOut())
			continue
		}
		if !isExportableMethod(method) {
			continue
		}

		toName := c.targetName(method.Name)
		if _, ok := to.Type().FieldByName(toName); ok {
			// log.Debugf("  -> func %q -> field totf: %v", method.Name, totf)
			tof := to.FieldByName(toName)
			// log.Debugf("  -> func %q -> field tof: %v", method.Name, tof)
			vom := from.Method(i)
			out := vom.Call([]reflect.Value{})
			tof.Set(out[0])
		}
	}
	return
}

func (c cloner) copyCloneableObject(z Cloneable, to Value, fromVar, toVar interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			f, t := ValueOf(fromVar), ValueOf(toVar)
			ft, tt := f.Type(), t.Type()
			fk, tk := ft.Kind(), tt.Kind()
			log.Debugf("error on setting %v (%v) -> %v (%v)", ft, fk, tt, tk)
			// log.Errorf("error on setting %q -> %q: %v", fromName, toName, e)
			if e2, ok := e.(error); ok {
				err = e2
			} else {
				err = errors.New("recovered cloneable error: %v", e)
			}
		}
	}()

	newI := z.Clone()
	var too, tol reflect.Value
	for too = to.Value; too.Kind() == reflect.Ptr; {
		tol = too
		too = too.Elem()
	}
	tol.Set(reflect.ValueOf(newI))
	// to.IndirectValueRecursive().Set(reflect.ValueOf(newI))
	return
}

func (c cloner) copyFieldToField(fromField reflect.Value, to held, oft, ott reflect.Type, fromName, toName string) (err error) {
	// fiv, tiv := fromField.IsValid(), toField.IsValid()
	// if !fiv || !tiv {
	//	return
	// }

	// if fromName == "DoubleAge" || fromName == "Role" {
	//	log.Debugf("%q (%v) -> %q (%v)", fromName, oft, toName, ott)
	// }

	//if !fromField.Type().AssignableTo(toField.Type()) {
	//	return // ignore copying between the different types if them can be assigned to the opposite one.
	//}

	toField := to.TargetField()
	fk, tk := fromField.Kind(), toField.Kind()

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("error on setting %q (%v,%v) -> %q (%v,%v): %v", fromName, oft, fk, toName, ott, tk, e)
			if e2, ok := e.(error); ok {
				err = e2
			} else {
				err = errors.New("recovered error: %v", e)
			}
		}
	}()

	if fk == reflect.Func {
		// func -> any
		// if fromName == "Role" {
		// 	log.Debugf("1. %v -> %v", oft, ott)
		// }
		if tk != reflect.Func {
			// func toField field
			// log.Debugf("1.1 %v -> %v", oft, ott)
			err = c.copyFuncToField(oft, ott, fromName, toName, fromField, toField)
		}
		return

	} else if tk == reflect.Func {
		// val -> func
		// if fromName == "Role" {
		// 	log.Debugf("2. %v -> %v: nothing needed toField do", oft, ott)
		// }
		if fk != reflect.Func {
			// func toField field
			// log.Debugf("2.1 %v -> %v", oft, ott)
			err = c.copyFieldToFunc(oft, ott, fromName, toName, fromField, toField)
		}
		return

		// } else if tk == reflect.Slice {
		// 	// note that array have not been target yet
		// 	newTargetType := toField.Type().Elem()
		// 	newTarget := reflect.New(newTargetType)
		// 	newTo := IndirectValue(newTarget)
		// 	if err = c.copyFieldToField(fromField, newTo, fromStructField, oft, ott, fromName, toName); err == nil {
		// 		if toField.Len() == 0 {
		// 			toField.Set(reflect.Append(toField, newTo))
		// 		} else {
		// 			toField.Index(0).Set(newTo)
		// 		}
		// 	}
		// 	return

	} else if tk == reflect.Ptr {
		// such as: string -> *string
		// log.Debugf("toField: %v (%v)", ott.Name(), ott)
		if !toField.IsValid() || fk != reflect.Ptr {
			typ := toField.Type()
			for typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
				// log.Debugf("  .. %v %v | toField: %v %v", typ.Kind(), typ, tk, toField)
			}
			// typ = fromField.Type()
			// for typ.Kind() == reflect.Ptr {
			//	typ = typ.Elem()
			// }
			toV := reflect.New(typ)
			toV.Elem().Set(fromField)
			toField.Set(toV)
			// log.Debugf("toV: %v (%v) = %v / %v", toField.Type().Name(), toField.Type(), toField.Pointer(), toField.Elem().Interface())
		} else if fk == reflect.Ptr {
			toField.Set(fromField)
		} else {
			err = errors.New("??? ??? error on setting %q (%v) -> %q (%v).", fromName, oft, toName, ott)
			// toField.IndirectValueRecursive().Set(fromField)
		}
		return
		// } else if fk == reflect.Ptr && tk == reflect.Ptr {
		//	// ptr -> ptr, ok, nothing needed toField do
		//	if fromName == "Birthday" {
		//		log.Debug("")
		//	}
		//	if !toField.Elem().IsValid() {
		//		toV := reflect.New(toField.Type().Elem())
		//		toV.Elem().Set(fromField)
		//		toField.Set(toV)
		//		// log.Debugf("toV: %v (%v) = %v / %v", toField.Type().Name(), toField.Type(), toField.Pointer(), toField.Elem().Interface())
		//	} else {
		//		// err = errors.New("??? ??? error on setting %q (%v) -> %q (%v).", fromName, fromField.Type(), toName, toField.Type())
		//		toField.Set(fromField)
		//	}
		//	// log.Debugf("3. %v -> %v", oft, ott)
		//	return
	} else if oft == ott {
		// copying between two equivalent types, ok, nothing needed toField do
		// log.Debugf("4. %v -> %v | %q", oft, ott, fromName)

	} else {
		// follow the pointer of object if necessary

		fField := IndirectValue(fromField)
		if !fField.IsValid() {
			fField = reflect.New(oft.Elem())
		}
		fromField = IndirectValue(fField)

		//if CanIsZero(toField) && IsZero(toField) {
		//	tField := IndirectValue(toField)
		//	if !tField.IsValid() {
		//		// log.Debugf("ott: %v %v | tField: %v %v", ott.Kind(), ott, tField.Kind(), tField.Type())
		//		tField = reflect.New(ott.Elem())
		//	}
		//	toField = IndirectValue(tField)
		//}
		//
		//// fiv, tiv = fromField.IsValid(), toField.IsValid()
		//// if !tiv {
		////	return
		//// }
		//if !toField.IsValid() {
		//	var tv interface{} = Value{toField}.GetValue()
		//	log.Debugf("target field ignored: %q (%v, ott=%v, tv=%v)", toName, tk, ott, tv)
		//	return
		//}
		oft = fromField.Type()
		fk = oft.Kind()
		if toField.IsValid() {
			ott = toField.Type()
			tk = ott.Kind()
		}
	}

	if to.CanSet() {
		if needReset := c.needReset(fromField, to); needReset {
			SetZero(toField)
		} else if canCopy, isNilOrZeroSkipped, cannotAssignTo := c.canCopy(fromField, to, oft, ott); canCopy || isNilOrZeroSkipped {
			if canCopy {
				// toField.Set(fromField)
				to.Set(fromField)
			} // else if isNilOrZeroSkipped { // nothing needed toField do }
		} else if cannotAssignTo {
			if (fk == reflect.Slice || fk == reflect.Array) && (tk == reflect.Slice || tk == reflect.Array) {
				if oft.Elem() == ott.Elem() {
					// such as: [2]string -> [3]string
					if oft.Len() < ott.Len() {
						for i := 0; i < fromField.Len(); i++ {
							toField.Index(i).Set(fromField.Index(i))
						}
					}
				} else {
					// such as: []int => []string
					err = errors.New("error on setting %q (%v) -> %q (%v): TODO: such as: []int => []string", fromName, oft, toName, ott)
				}
			} else if isKindInt(fk) && isKindInt(tk) {
				i := fromField.Int()
				toField.SetInt(i)
			} else if isKindFloat(fk) && isKindFloat(tk) {
				i := fromField.Float()
				toField.SetFloat(i)
			} else if isKindComplex(fk) && isKindComplex(tk) {
				i := fromField.Complex()
				toField.SetComplex(i)
			} else {
				err = errors.New("error on setting %q (%v) -> %q (%v): canCopy=%v|isNilOrZeroSkipped=%v|cannotAssignTo=%v, needReset=%v", fromName, oft, toName, ott, canCopy, isNilOrZeroSkipped, cannotAssignTo, needReset)
			}
		} else {
			err = errors.New("??? ??? error on setting %q (%v) -> %q (%v): canCopy=%v|isNilOrZeroSkipped=%v|cannotAssignTo=%v, needReset=%v", fromName, oft, toName, ott, canCopy, isNilOrZeroSkipped, cannotAssignTo, needReset)
		}
	} else if tk == reflect.Ptr {
		// any (non-ptr) -> ptr
		toField = IndirectValue(toField)
	} else {
		var fv, tv interface{}
		if fromField.CanInterface() {
			fv = fromField.Interface()
		}
		if toField.CanInterface() {
			tv = toField.Interface()
		}
		err = errors.New("target cannot be set: field %q (%v) -> %q (%v) (value: src=%v, tgt=%v)", fromName, oft, toName, ott, fv, tv)
	}
	return
}

func (c cloner) copyFuncToField(ft, tt reflect.Type, fromName, toName string, fromField, toField reflect.Value) (err error) {
	if ft.NumIn() == 0 && ft.NumOut() == 1 {
		out := fromField.Call([]reflect.Value{})
		toField.Set(out[0])
	}
	return
}

func (c cloner) copyFieldToFunc(ft, tt reflect.Type, fromName, toName string, fromField, toField reflect.Value) (err error) {
	if tt.NumIn() == 1 {
		_ = toField.Call([]reflect.Value{fromField})
	}
	return
}

func (c cloner) canCopy(from reflect.Value, to held, fromType, toType reflect.Type) (canCopy, isNilOrZeroSkipped, cannotAssignTo bool) {
	tof := to.TargetField()
	if from.Kind() != tof.Kind() {
		if fromType.AssignableTo(toType) {
			canCopy = true
		} else {
			cannotAssignTo = true
		}
		return
	}
	if fromType.AssignableTo(toType) {
		if c.EachFieldAlways {
			canCopy = true
		} else if CanIsNil(from) {
			if IsNil(from) {
				canCopy, isNilOrZeroSkipped = c.KeepIfFromIsNil, true
			} else {
				canCopy = true
			}
		} else if CanIsZero(from) {
			if IsZero(from) {
				canCopy, isNilOrZeroSkipped = c.KeepIfFromIsZero, true
			} else {
				canCopy = true
			}
		}
	} else {
		cannotAssignTo = true
	}
	return
}

func (c cloner) needReset(from reflect.Value, to held) (needReset bool) {
	tof := to.TargetField()
	if from.Kind() != tof.Kind() {
		return
	}
	if Equal(from, tof) {
		needReset = c.ZeroIfEqualsFrom
	}
	return
}

func (c cloner) targetName(fromName string) (toName string) {
	var mapped bool
	var toNameTmp string
	toName = fromName
	if c.nameMappingRule != nil {
		toNameTmp, mapped = c.nameMappingRule(fromName)
		if mapped {
			toName = toNameTmp
		}
	}
	if !mapped {
		if z, ok := c.nameMappingMap[fromName]; ok {
			toName = z
		}
	}
	return
}

func (c cloner) shouldBeIgnored(name string) (ignored bool) {
	for _, n := range c.ignoredNames {
		if name == n {
			ignored = true
			break
		}
	}
	return
}

// isSeparator reports whether the rune could mark a word boundary.
func (c cloner) isSeparator(r rune) bool { return isSeparator(r) }

// Captalize returns a copy of the string s with all Unicode letters that begin words
// mapped to their Unicode title case.
//
// BUG(rsc): The rule Title uses for word boundaries does not handle Unicode punctuation properly.
func (c cloner) Captalize(s string) string {
	// Use a closure here to remember state.
	// Hackish but effective. Depends on Map scanning in order and calling
	// the closure once per rune.
	prev := ' '
	return strings.Map(
		func(r rune) rune {
			if c.isSeparator(prev) {
				prev = r
				return unicode.ToTitle(r)
			}
			prev = r
			return r
		},
		s)
}

func (c cloner) tryConvert(v reflect.Value, t reflect.Type) (out reflect.Value, err error) {
	defer func() {
		if e := recover(); e != nil {
			if e2, ok := e.(error); ok {
				err = e2
			} else {
				err = errors.New("%v", e)
			}
		}
	}()

	out = v.Convert(t)
	return
}

func (c cloner) indirectCreate(fromType reflect.Type) (newTargetType reflect.Type, parent, newTo Value) {
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
	newTargetType, tp, newTo = c.indirectCreate(tt)
	parent = tp
	if tp.CanAddr() {
		parent = ValueOf(reflect.New(fromType))
		parent.Set(tp.Addr())
	}
	return
}
