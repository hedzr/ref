package ref

import (
	"github.com/hedzr/assert"
	"reflect"
	"testing"
)

type testStruct struct {
	unexported uint64
	Dummy      string `test:"dummytag"`
	Yummy      int    `test:"yummytag"`
}

func TestMisc(t *testing.T) {
	defer initLogger(t)()

	t.Run("for TypeOf", testTypeOf)
	t.Run("for FieldDeeper", testFieldDeeper)
	t.Run("for FieldTagDeeper", testFieldTagDeeper)
	t.Run("testUnexported", testUnexported)
}

type Foo struct {
	Exported   string
	unexported string
}

func testUnexported(t *testing.T) {
	f := &Foo{
		Exported: "Old Value ",
	}

	t.Log(f.Exported)

	field := reflect.ValueOf(f).Elem().FieldByName("unexported")
	SetUnexportedField(field, "New Value")
	t.Log(GetUnexportedField(field))
	t.Logf("foo: %+v", f)
}

func testTypeOf(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	typ := reflect.TypeOf(dummyStruct)
	t.Logf("typ: %+v", typ)

	if fld, ok := typ.FieldByName("Dummy"); ok {
		if tag, ok := fld.Tag.Lookup("test"); ok {
			t.Logf("tag of Dummy: %q", tag)
		}
	}
}

func testFieldDeeper(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	err := FieldsDeeper(dummyStruct, func(owner *reflect.Value, thisField reflect.StructField, this reflect.Value) {
		t.Logf("    -> owner: %v, this: %v / %v", owner, thisField, this)
	})
	assert.NoError(t, err)
}

func testFieldTagDeeper(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	err := FieldsTagDeeper(dummyStruct, "test", func(owner *reflect.Value, thisField reflect.StructField, this reflect.Value, tagValue string) {
		t.Logf("    -> owner: %v, this: %v / %v, tag-value: %q", owner, thisField, this, tagValue)
	})
	assert.NoError(t, err)
}

func TestGetField(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testGetFieldOnStruct)
	t.Run("on struct pointer", testGetFieldOnStructPointer)
	t.Run("on non-struct", testGetFieldOnNonStruct)
	t.Run("on non-existing-field", testGetFieldNonExistingField)
	t.Run("on unexported field", testGetFieldUnexportedField)
}

func testGetFieldOnStruct(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	value, err := GetField(dummyStruct, "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, value, "test")
}

func testGetFieldOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
	}

	value, err := GetField(dummyStruct, "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, value, "test")
}

func testGetFieldOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := GetField(dummy, "Dummy")
	assert.Error(t, err)
}

func testGetFieldNonExistingField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	_, err := GetField(dummyStruct, "obladioblada")
	assert.Error(t, err)
}

func testGetFieldUnexportedField(t *testing.T) {
	dummyStruct := testStruct{
		unexported: 12345,
		Dummy:      "test",
	}

	assert.PanicMatches(t, func() {
		GetField(dummyStruct, "unexported")
	}, "reflect.Value.Interface: cannot return value obtained from unexported field or method")
}

func TestGetFieldKind(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testGetFieldKindOnStruct)
	t.Run("on struct pointer", testGetFieldKindOnStructPointer)
	t.Run("on non-struct", testGetFieldKindOnNonStruct)
	t.Run("on non-existing field", testGetFieldKindNonExistingField)
}

func testGetFieldKindOnStruct(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	kind, err := GetFieldKind(dummyStruct, "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, kind, reflect.String)

	kind, err = GetFieldKind(dummyStruct, "Yummy")
	assert.NoError(t, err)
	assert.Equal(t, kind, reflect.Int)
}

func testGetFieldKindOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	kind, err := GetFieldKind(dummyStruct, "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, kind, reflect.String)

	kind, err = GetFieldKind(dummyStruct, "Yummy")
	assert.NoError(t, err)
	assert.Equal(t, kind, reflect.Int)
}

func testGetFieldKindOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := GetFieldKind(dummy, "Dummy")
	assert.Error(t, err)
}

func testGetFieldKindNonExistingField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	_, err := GetFieldKind(dummyStruct, "obladioblada")
	assert.Error(t, err)
}

func TestGetFieldType(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testGetFieldTypeOnStruct)
	t.Run("on struct pointer", testGetFieldTypeOnStructPointer)
	t.Run("on non-struct", testGetFieldTypeOnNonStruct)
	t.Run("on non-existing field", testGetFieldTypeNonExistingField)
}

func testGetFieldTypeOnStruct(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	typeString, err := GetFieldType(dummyStruct, "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, typeString, "string")

	typeString, err = GetFieldType(dummyStruct, "Yummy")
	assert.NoError(t, err)
	assert.Equal(t, typeString, "int")
}

func testGetFieldTypeOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	typeString, err := GetFieldType(dummyStruct, "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, typeString, "string")

	typeString, err = GetFieldType(dummyStruct, "Yummy")
	assert.NoError(t, err)
	assert.Equal(t, typeString, "int")
}

func testGetFieldTypeOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := GetFieldType(dummy, "Dummy")
	assert.Error(t, err)
}

func testGetFieldTypeNonExistingField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	_, err := GetFieldType(dummyStruct, "obladioblada")
	assert.Error(t, err)
}

func TestGetFieldTag(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testGetFieldTagOnStruct)
	t.Run("on struct pointer", testGetFieldTagOnStructPointer)
	t.Run("on non-struct", testGetFieldTagOnNonStruct)
	t.Run("on non-existing field", testGetFieldTagNonExistingField)
	t.Run("on unexported field", testGetFieldTagUnexportedField)
}

func testGetFieldTagOnStruct(t *testing.T) {
	dummyStruct := testStruct{}

	tag, err := GetFieldTag(dummyStruct, "test", "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, tag, "dummytag")

	tag, err = GetFieldTag(dummyStruct, "test", "Yummy")
	assert.NoError(t, err)
	assert.Equal(t, tag, "yummytag")
}

func testGetFieldTagOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{}

	tag, err := GetFieldTag(dummyStruct, "test", "Dummy")
	assert.NoError(t, err)
	assert.Equal(t, tag, "dummytag")

	tag, err = GetFieldTag(dummyStruct, "test", "Yummy")
	assert.NoError(t, err)
	assert.Equal(t, tag, "yummytag")
}

func testGetFieldTagOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := GetFieldTag(dummy, "test", "Dummy")
	assert.Error(t, err)
}

func testGetFieldTagNonExistingField(t *testing.T) {
	dummyStruct := testStruct{}

	_, err := GetFieldTag(dummyStruct, "test", "obladioblada")
	assert.Error(t, err)
}

func testGetFieldTagUnexportedField(t *testing.T) {
	dummyStruct := testStruct{
		unexported: 12345,
		Dummy:      "test",
	}

	_, err := GetFieldTag(dummyStruct, "test", "unexported")
	assert.Error(t, err)
}

func TestSetField(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct with valid value type", testSetFieldOnStructWithValidValueType)
	t.Run("on non-existing field", testSetFieldNonExistingField)
	t.Run("on invalid value type", testSetFieldInvalidValueType)
	t.Run("on non-exported field", testSetFieldNonExportedField)
}

func testSetFieldOnStructWithValidValueType(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	err := SetField(&dummyStruct, "Dummy", "abc")
	assert.NoError(t, err)
	assert.Equal(t, dummyStruct.Dummy, "abc")
}

// func TestSetField_on_non_struct(t *testing.T) {
//     dummy := "abc 123"

//     err := SetField(&dummy, "Dummy", "abc")
//     assert.NoError(t, err)
// }

func testSetFieldNonExistingField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	err := SetField(&dummyStruct, "obladioblada", "life goes on")
	assert.Error(t, err)
}

func testSetFieldInvalidValueType(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	err := SetField(&dummyStruct, "Yummy", "123")
	assert.Error(t, err)
}

func testSetFieldNonExportedField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
	}

	assert.Error(t, SetField(&dummyStruct, "unexported", "fail, bitch"))
}

func TestFields(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testFieldsOnStruct)
	t.Run("on struct pointer", testFieldsOnStructPointer)
	t.Run("on non-struct", testFieldsOnNonStruct)
	t.Run("on non-exported fields", testFieldsWithNonExportedFields)
}

func testFieldsOnStruct(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	fields, err := Fields(dummyStruct)
	assert.NoError(t, err)
	assert.Equal(t, fields, []string{"Dummy", "Yummy"})
}

func testFieldsOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	fields, err := Fields(dummyStruct)
	assert.NoError(t, err)
	assert.Equal(t, fields, []string{"Dummy", "Yummy"})
}

func testFieldsOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := Fields(dummy)
	assert.Error(t, err)
}

func testFieldsWithNonExportedFields(t *testing.T) {
	dummyStruct := testStruct{
		unexported: 6789,
		Dummy:      "test",
		Yummy:      123,
	}

	fields, err := Fields(dummyStruct)
	assert.NoError(t, err)
	assert.Equal(t, fields, []string{"Dummy", "Yummy"})
}

func TestHasField(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct with existing field", testHasFieldOnStructWithExistingField)
	t.Run("on struct pointer with existing field", testHasFieldOnStructPointerWithExistingField)
	t.Run("on non-existing field", testHasFieldNonExistingField)
	t.Run("on non-struct", testHasFieldOnNonStruct)
	t.Run("on unexported field", testHasFieldUnexportedField)
}

func testHasFieldOnStructWithExistingField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	has := HasField(dummyStruct, "Dummy")
	assert.Equal(t, has, true)
}

func testHasFieldOnStructPointerWithExistingField(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	has := HasField(dummyStruct, "Dummy")
	assert.Equal(t, has, true)
}

func testHasFieldNonExistingField(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	has := HasField(dummyStruct, "Test")
	assert.Equal(t, has, false)
}

func testHasFieldOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	has := HasField(dummy, "Test")
	assert.Equal(t, has, false)
}

func testHasFieldUnexportedField(t *testing.T) {
	dummyStruct := testStruct{
		unexported: 7890,
		Dummy:      "test",
		Yummy:      123,
	}

	has := HasField(dummyStruct, "unexported")
	assert.Equal(t, has, false)
}

func TestTags(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testTagsOnStruct)
	t.Run("on struct pointer", testTagsOnStructPointer)
	t.Run("on non-struct", testTagsOnNonStruct)
}

func testTagsOnStruct(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	tags, err := Tags(dummyStruct, "test")
	assert.NoError(t, err)
	assert.Equal(t, tags, map[string]string{
		"Dummy": "dummytag",
		"Yummy": "yummytag",
	})
}

func testTagsOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	tags, err := Tags(dummyStruct, "test")
	assert.NoError(t, err)
	assert.Equal(t, tags, map[string]string{
		"Dummy": "dummytag",
		"Yummy": "yummytag",
	})
}

func testTagsOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := Tags(dummy, "test")
	assert.Error(t, err)
}

func TestItems(t *testing.T) {
	defer initLogger(t)()

	t.Run("on struct", testItemsOnStruct)
	t.Run("on struct pointer", testItemsOnStructPointer)
	t.Run("on non-struct", testItemsOnNonStruct)
}

func testItemsOnStruct(t *testing.T) {
	dummyStruct := testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	tags, err := Items(dummyStruct)
	assert.NoError(t, err)
	assert.Equal(t, tags, map[string]interface{}{
		"Dummy": "test",
		"Yummy": 123,
	})
}

func testItemsOnStructPointer(t *testing.T) {
	dummyStruct := &testStruct{
		Dummy: "test",
		Yummy: 123,
	}

	tags, err := Items(dummyStruct)
	assert.NoError(t, err)
	assert.Equal(t, tags, map[string]interface{}{
		"Dummy": "test",
		"Yummy": 123,
	})
}

func testItemsOnNonStruct(t *testing.T) {
	dummy := "abc 123"

	_, err := Items(dummy)
	assert.Error(t, err)
}

func TestDeep(t *testing.T) {
	defer initLogger(t)()

	t.Run("items deep", testItemsDeep)
	t.Run("tags deep", testTagsDeep)
	t.Run("fields deep", testFieldsDeep)
}

func testItemsDeep(t *testing.T) {
	type Address struct {
		Street string `tag:"be"`
		Number int    `tag:"bi"`
	}

	type unexportedStruct struct{}

	type Person struct {
		Name string `tag:"bu"`
		Address
		unexportedStruct
	}

	p := Person{}
	p.Name = "John"
	p.Street = "Decumanus maximus"
	p.Number = 17

	items, err := Items(p)
	assert.NoError(t, err)
	itemsDeep, err := ItemsDeep(p)
	assert.NoError(t, err)

	assert.Equal(t, len(items), 2)
	assert.Equal(t, len(itemsDeep), 3)
	assert.Equal(t, itemsDeep["Name"], "John")
	assert.Equal(t, itemsDeep["Street"], "Decumanus maximus")
	assert.Equal(t, itemsDeep["Number"], 17)
}

func testTagsDeep(t *testing.T) {
	type Address struct {
		Street string `tag:"be"`
		Number int    `tag:"bi"`
	}

	type unexportedStruct struct{}

	type Person struct {
		Name string `tag:"bu"`
		Address
		unexportedStruct
	}

	p := Person{}
	p.Name = "John"
	p.Street = "Decumanus maximus"
	p.Number = 17

	tags, err := Tags(p, "tag")
	assert.NoError(t, err)
	tagsDeep, err := TagsDeep(p, "tag")
	assert.NoError(t, err)

	assert.Equal(t, len(tags), 2)
	assert.Equal(t, len(tagsDeep), 3)
	assert.Equal(t, tagsDeep["Name"], "bu")
	assert.Equal(t, tagsDeep["Street"], "be")
	assert.Equal(t, tagsDeep["Number"], "bi")
}

func testFieldsDeep(t *testing.T) {
	type Address struct {
		Street string `tag:"be"`
		Number int    `tag:"bi"`
	}

	type unexportedStruct struct{}

	type Person struct {
		Name string `tag:"bu"`
		Address
		unexportedStruct
	}

	p := Person{}
	p.Name = "John"
	p.Street = "street?"
	p.Number = 17

	fields, err := Fields(p)
	assert.NoError(t, err)
	fieldsDeep, err := FieldsDeep(p)
	assert.NoError(t, err)

	assert.Equal(t, len(fields), 2)
	assert.Equal(t, len(fieldsDeep), 3)
	assert.Equal(t, fieldsDeep[0], "Name")
	assert.Equal(t, fieldsDeep[1], "Street")
	assert.Equal(t, fieldsDeep[2], "Number")
}
