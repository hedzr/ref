package ref

import (
	"fmt"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
)

// GetField returns the value of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetField(obj interface{}, fieldName string) (field interface{}, err error) {
	var fields []interface{}
	fields, err = GetFields(obj, fieldName)
	if len(fields) > 0 {
		field = fields[0]
	}
	return
}

// GetFields returns the value of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetFields(obj interface{}, fieldNames ...string) (fields []interface{}, err error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		err = errors.New("Cannot use GetField on a non-struct interface")
		return
	}

	objValue := reflectValue(obj)
	for _, name := range fieldNames {
		field := objValue.FieldByName(name)
		if !field.IsValid() {
			err = errors.New("no such field: %s in object %v", name, obj)
			return
		}
		fields = append(fields, field.Interface())
	}
	return
}

// GetFieldKind returns the kind of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetFieldKind(obj interface{}, fieldName string) (kind reflect.Kind, err error) {
	var kinds []reflect.Kind
	kinds, err = GetFieldKinds(obj, fieldName)
	if len(kinds) > 0 {
		kind = kinds[0]
	}
	return
}

// GetFieldKind returns the kind of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetFieldKinds(obj interface{}, fieldNames ...string) (kinds []reflect.Kind, err error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		err = errors.New("Cannot use GetField on a non-struct interface")
		return
	}

	objValue := reflectValue(obj)
	for _, name := range fieldNames {
		field := objValue.FieldByName(name)
		if !field.IsValid() {
			err = errors.New("no such field: %s in object %v", name, obj)
			return
		}
		kinds = append(kinds, field.Type().Kind())
	}
	return
}

// GetFieldType returns the kind of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetFieldType(obj interface{}, fieldName string) (typ string, err error) {
	var types []string
	types, err = GetFieldTypes(obj, fieldName)
	if len(types) > 0 {
		typ = types[0]
	}
	return
}

// GetFieldType returns the kind of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetFieldTypes(obj interface{}, fieldNames ...string) (types []string, err error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		err = errors.New("Cannot use GetField on a non-struct interface")
		return
	}

	objValue := reflectValue(obj)
	for _, name := range fieldNames {
		field := objValue.FieldByName(name)
		if !field.IsValid() {
			err = errors.New("no such field: %s in object %v", name, obj)
			return
		}
		types = append(types, field.Type().String())
	}
	return
}

// GetFieldTag1 returns the provided obj field tag value. obj can whether
// be a structure or pointer to structure.
func GetFieldTag1(obj interface{}, tagKey string, fieldName string) (tagValue string, err error) {
	var tagValues []string
	tagValues, err = GetFieldTags(obj, tagKey, fieldName)
	if len(tagValues) > 0 {
		tagValue = tagValues[0]
	}
	return
}

// GetFieldTag returns the provided obj field tag value. obj can whether
// be a structure or pointer to structure.
func GetFieldTags(obj interface{}, tagKey string, fieldNames ...string) (tagValues []string, err error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		err = errors.New("Cannot use GetField on a non-struct interface")
		return
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	for _, name := range fieldNames {
		field, ok := objType.FieldByName(name)
		if !ok {
			err = errors.New("no such field: %s in object %v", name, obj)
			return
		}
		if !isExportableField(field) {
			err = errors.New("Cannot GetFieldTag on a non-exported struct field")
			return
		}

		tagValues = append(tagValues, field.Tag.Get(tagKey))
	}
	return
}

// SetField sets the provided obj field with provided value. obj param has
// to be a pointer to a struct, otherwise it will soundly fail. Provided
// value type should match with the struct field you're trying to set.
func SetField(obj interface{}, name string, value interface{}) (err error) {
	// Fetch the field reflect.Value
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		err = errors.New("no such field: %s in obj", name)
		return
	}

	// If obj field value is not settable an error is thrown
	if !structFieldValue.CanSet() {
		err = errors.New("cannot set %s field value", name)
		return
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		err = errors.New("provided value type didn't match obj field type")
		return
	}

	structFieldValue.Set(val)
	return
}

// HasField checks if the provided field name is part of a struct. obj can whether
// be a structure or pointer to structure.
func HasField(obj interface{}, name string) (has bool) {
	return HasAnyField(obj, name)
}

func HasAnyField(obj interface{}, names ...string) (has bool) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		return
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	for _, name := range names {
		field, ok := objType.FieldByName(name)
		if !ok || !isExportableField(field) {
			continue
		}
		has = ok
		break
	}
	return
}

// Fields returns the struct fields names list. obj can whether
// be a structure or pointer to structure.
func Fields(obj interface{}) ([]string, error) {
	return fields(obj, false)
}

// FieldsDeeper returns "flattened" fields (fields from anonymous
// inner structs are treated as normal fields)
func FieldsDeeper(obj interface{}, fn func(owner *reflect.Value, thisField reflect.StructField, this reflect.Value)) error {
	return fieldsDeeper(obj, fn, true, nil)
}

func fieldsDeeper(obj interface{}, fn func(owner *reflect.Value, thisField reflect.StructField, this reflect.Value), deep bool, owner *reflect.Value) (err error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		return errors.New("Cannot use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	for i := 0; i < fieldsCount; i++ {
		field := objType.Field(i)
		if isExportableField(field) {
			if deep && field.Anonymous {
				fieldValue := objValue.Field(i)
				fn(owner, field, fieldValue)
				err1 := fieldsDeeper(fieldValue.Interface(), fn, deep, &fieldValue)
				if err1 != nil {
					err = errors.New("cannot get fields in %s", field.Name).Attach(err)
				}
				// allFields = append(allFields, subFields...)
			} else {
				// allFields = append(allFields, field.Name)
				fieldValue := objValue.Field(i)
				fn(owner, field, fieldValue)
			}
		}
	}
	return
}

// FieldsDeep returns "flattened" fields (fields from anonymous
// inner structs are treated as normal fields)
func FieldsDeep(obj interface{}) ([]string, error) {
	return fields(obj, true)
}

func fields(obj interface{}, deep bool) ([]string, error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		return nil, errors.New("Cannot use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	var allFields []string
	for i := 0; i < fieldsCount; i++ {
		field := objType.Field(i)
		if isExportableField(field) {
			if deep && field.Anonymous {
				fieldValue := objValue.Field(i)
				subFields, err := fields(fieldValue.Interface(), deep)
				if err != nil {
					return nil, fmt.Errorf("Cannot get fields in %s: %s", field.Name, err.Error())
				}
				allFields = append(allFields, subFields...)
			} else {
				allFields = append(allFields, field.Name)
			}
		}
	}

	return allFields, nil
}

// Items returns the field - value struct pairs as a map. obj can whether
// be a structure or pointer to structure.
func Items(obj interface{}) (map[string]interface{}, error) {
	return items(obj, false)
}

// FieldsDeep returns "flattened" items (fields from anonymous
// inner structs are treated as normal fields)
func ItemsDeep(obj interface{}) (map[string]interface{}, error) {
	return items(obj, true)
}

func items(obj interface{}, deep bool) (map[string]interface{}, error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		return nil, errors.New("Cannot use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	allItems := make(map[string]interface{})

	for i := 0; i < fieldsCount; i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i)
		if isExportableField(field) {
			if deep && field.Anonymous {
				if m, err := items(fieldValue.Interface(), deep); err == nil {
					for k, v := range m {
						allItems[k] = v
					}
				} else {
					return nil, fmt.Errorf("Cannot get items in %s: %s", field.Name, err.Error())
				}
			} else {
				allItems[field.Name] = fieldValue.Interface()
			}
		}
	}

	return allItems, nil
}

// Tags lists the struct tag fields. obj can whether
// be a structure or pointer to structure.
func Tags(obj interface{}, key string) (map[string]string, error) {
	return tags(obj, key, false)
}

// FieldsDeep returns "flattened" tags (fields from anonymous
// inner structs are treated as normal fields)
func TagsDeep(obj interface{}, key string) (map[string]string, error) {
	return tags(obj, key, true)
}

func tags(obj interface{}, key string, deep bool) (map[string]string, error) {
	if !hasAnyValidTypes(obj, reflect.Struct, reflect.Ptr) {
		return nil, errors.New("Cannot use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	allTags := make(map[string]string)

	for i := 0; i < fieldsCount; i++ {
		structField := objType.Field(i)
		if isExportableField(structField) {
			if deep && structField.Anonymous {
				fieldValue := objValue.Field(i)
				if m, err := tags(fieldValue.Interface(), key, deep); err == nil {
					for k, v := range m {
						allTags[k] = v
					}
				} else {
					return nil, fmt.Errorf("Cannot get items in %s: %s", structField.Name, err.Error())
				}
			} else {
				allTags[structField.Name] = structField.Tag.Get(key)
			}
		}
	}

	return allTags, nil
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

func isExportableField(field reflect.StructField) bool {
	// PkgPath is empty for exported fields.
	return field.PkgPath == ""
}

func hasAnyValidTypes(obj interface{}, types ...reflect.Kind) bool {
	for _, t := range types {
		if reflect.TypeOf(obj).Kind() == t {
			return true
		}
	}
	return false
}

func hasAllValidTypes(obj interface{}, types ...reflect.Kind) bool {
	for _, t := range types {
		if reflect.TypeOf(obj).Kind() != t {
			return false
		}
	}
	return true
}

func isStruct(obj interface{}) bool {
	return reflect.TypeOf(obj).Kind() == reflect.Struct
}

func isPointer(obj interface{}) bool {
	return reflect.TypeOf(obj).Kind() == reflect.Ptr
}
