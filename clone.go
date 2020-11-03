package ref

import (
	"bytes"
	"encoding/gob"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
)

type (
	// Cloner provides a copier interface for the deep clone algor.
	Cloner interface {
		Copy(fromVar, toVar interface{}, Opts ...Opt) (err error)
	}

	// Opt is functional option functor for Cloner
	Opt func(c Cloner)

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

	DefaultCloner = cloner{
		ignoredNames:          nil,
		nameMappingMap:        make(map[string]string),
		nameMappingRule:       DefaultNameMappingRule,
		IgnoreUnexportedError: true,
		KeepIfFromIsNil:       false,
		ZeroIfEqualsFrom:      false,
		KeepIfFromIsZero:      false,
		EachFieldAlways:       false,
	}

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
func Clone(fromVar, toVar interface{}) interface{} {
	_ = DefaultCloner.Copy(fromVar, toVar)
	return toVar
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
	KeepIfFromIsNil       bool // 源字段值为nil指针时，目标字段的值保持不变
	ZeroIfEqualsFrom      bool // 源和目标字段值相同时，目标字段被清除为未初始化的零值
	KeepIfFromIsZero      bool // 源字段值为未初始化的零值时，目标字段的值保持不变 // 此条尚未实现
	EachFieldAlways       bool
}

func (c cloner) Copy(fromVar, toVar interface{}, Opts ...Opt) (err error) {
	if !hasAnyValidTypes(fromVar, reflect.Struct, reflect.Ptr) {
		err = errors.New("fromVar should be a valid struct object or pointer object")
		return
	}
	if !hasAnyValidTypes(toVar, reflect.Struct, reflect.Ptr) {
		err = errors.New("toVar should be a valid struct object or pointer object")
		return
	}

	from := reflectValue(fromVar)
	to := reflectValue(toVar)
	objType := from.Type()

	//modelType := reflect.TypeOf((*Cloneable)(nil)).Elem()
	//if objType.Implements(modelType) {
	//	//
	//}
	if z, ok := fromVar.(Cloneable); ok {
		newI := z.Clone()
		to.Set(reflect.ValueOf(newI))
		return
	}

	if !to.CanAddr() {
		return errors.New("clone to value is unaddressable")
	}

	if !from.IsValid() {
		return errors.New("clone to value is unaddressable")
	}

	if CanIsNil(from) && IsNil(from) {
		SetNil(to)
		return
	}

	if CanIsZero(from) && IsZero(from) {
		SetZero(to)
		return
	}

	fieldsCount := objType.NumField()
	for i := 0; i < fieldsCount; i++ {
		field := objType.Field(i)
		if c.shouldBeIgnored(field.Name) {
			continue
		}

		if isExportableField(field) {
			if field.Anonymous {
				//fieldValue := from.Field(i)
				//subFields, err := fields(fieldValue.Interface(), true)
				//if err != nil {
				//	return nil, fmt.Errorf("Cannot get fields in %s: %s", field.Name, err.Error())
				//}
				err = errors.New("nested structure on field %q", field.Name)
				return
			} else {
				vov := from.Field(i)
				tof := to.FieldByName(c.targetName(field.Name))
				if tof.CanSet() {
					tof.Set(vov)
				} else {
					err = errors.New("target cannot be set: field %q (value: %v)", field.Name, vov.Interface())
					return
				}
			}
		} else {
			vov := from.Field(i)
			tof := to.FieldByName(c.targetName(field.Name))
			if tof.CanSet() {
				tof.Set(vov)
			} else if !c.IgnoreUnexportedError {
				err = errors.New("target cannot be set: field %q (value: %v)", field.Name, vov.Interface())
				return
			}
		}
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
