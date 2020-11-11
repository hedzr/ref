package ref

import (
	"fmt"
	"reflect"
)

type held interface {
	// Get returns the reflect.Value object of a field from its name
	Get(sourceField reflect.StructField, toName string) reflect.Value

	CanSet() bool
	// Set can be called after first Get(name) invoked
	Set(newVal reflect.Value)
	// GetSourceField can be called after first Get(name) invoked
	GetSourceField() reflect.StructField
	// OriginalTarget returns the original 'to' reflect.Value object
	OriginalTarget() reflect.Value
	SetTargetField(tof reflect.Value)
	// TargetField returns the final 'tof' reflect.Value object (target of field)
	TargetField() reflect.Value
}

func newHeldStruct(to reflect.Value) held {
	h := &heldStruct{targetObj: to}
	switch h.targetObj.Kind() {
	case reflect.Struct:
	case reflect.Map:
		return newHeldMap(to)
	default:
		panic(fmt.Sprintf("want a struct object, but pass with %v", to.Type()))
	}
	if CanIsNil(to) && IsNil(to) { // struct is nil?
		// if an empty struct found, create a new one
		nmi := reflect.New(to.Type())
		to.Addr().Elem().Set(nmi)
	}
	return h
}

type heldStruct struct {
	sourceField reflect.StructField
	targetObj   reflect.Value
	fieldName   string
	field       reflect.Value
	tof         reflect.Value
}

func (h *heldStruct) FromStruct(to reflect.Value)              { h.targetObj = to }
func (h *heldStruct) OriginalTarget() (to reflect.Value)       { return h.targetObj }
func (h *heldStruct) SetTargetField(tof reflect.Value)         { h.tof = tof }
func (h *heldStruct) TargetField() (to reflect.Value)          { return h.tof }
func (h *heldStruct) CanSet() bool                             { return h.field.CanSet() }
func (h *heldStruct) Set(newVal reflect.Value)                 { h.field.Set(newVal) }
func (h *heldStruct) GetSourceField() (sf reflect.StructField) { return h.sourceField }
func (h *heldStruct) Get(sourceField reflect.StructField, toName string) reflect.Value {
	h.sourceField, h.fieldName = sourceField, toName
	h.field = h.targetObj.FieldByName(h.fieldName)
	return h.field
}

func newHeldMap(to reflect.Value) held {
	h := &heldMap{targetObj: to}
	switch h.targetObj.Kind() {
	case reflect.Struct:
		return newHeldStruct(to)
	case reflect.Map:
	default:
		panic(fmt.Sprintf("want a struct object, but pass with %v", to.Type()))
	}
	if CanIsNil(to) && IsNil(to) { // map is nil?
		// if an empty map found, create a new one
		nmi := reflect.MakeMap(to.Type())
		to.Addr().Elem().Set(nmi)
	}
	return h
}

type heldMap struct {
	sourceField reflect.StructField
	targetObj   reflect.Value
	keyName     reflect.Value
	tof         reflect.Value
}

func (h *heldMap) FromMap(to reflect.Value)                 { h.targetObj = to }
func (h *heldMap) OriginalTarget() (to reflect.Value)       { return h.targetObj }
func (h *heldMap) SetTargetField(tof reflect.Value)         { h.tof = tof }
func (h *heldMap) TargetField() (to reflect.Value)          { return h.tof }
func (h *heldMap) CanSet() bool                             { return true }
func (h *heldMap) Set(newVal reflect.Value)                 { h.targetObj.SetMapIndex(h.keyName, newVal) }
func (h *heldMap) GetSourceField() (sf reflect.StructField) { return h.sourceField }
func (h *heldMap) Get(sourceField reflect.StructField, toName string) reflect.Value {
	h.sourceField = sourceField
	h.keyName = reflect.ValueOf(toName)
	return h.targetObj.MapIndex(h.keyName)
}
