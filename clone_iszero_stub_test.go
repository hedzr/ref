package ref

import (
	"github.com/hedzr/assert"
	"reflect"
	"testing"
	"unsafe"
)

func TestCloneIsZero(t *testing.T) {
	t.Run("reflect set struct field", testSetStructField)

	t.Run("SetZero", testSetZero)
	t.Run("SetNil", testSetNil)
	t.Run("SetFieldValueUnsafe", testSetFieldValueUnsafe)
}

func testSetFieldValueUnsafe(t *testing.T) {
	cat := &Cat{
		Age:     9,
		name:    "cat",
		friends: []string{},
	}

	v := reflect.ValueOf(cat).Elem()
	v.FieldByName("Age").SetInt(11)

	type VV struct {
		typ  unsafe.Pointer
		ptr  unsafe.Pointer
		flag uintptr
	}

	v2 := (*VV)(unsafe.Pointer(&v))
	println("v2.ptr: ", v2.ptr)

	type CatX struct {
		Age     int
		Name    string
		friends []string
	}

	c2 := (*CatX)(unsafe.Pointer(cat))
	c2.Name = "ohmygod"

	t.Logf("cat  : %+v", cat)
	t.Logf("cat 2: %+v", c2)
}

func testSetZero(t *testing.T) {
	u1 := U{Name: "11"}
	v := reflect.ValueOf(&u1).Elem()
	fld := v.FieldByName("Name")
	SetZero(fld)
	t.Logf("after SetZero, v is: %+v", v)
}

func testSetNil(t *testing.T) {
	u1 := U{Name: "11", Birthday: &now}
	v := reflect.ValueOf(&u1).Elem()
	fld := v.FieldByName("Birthday")
	SetNil(fld)
	t.Logf("after SetNil, v is: %+v", v)
}

func testSetStructField(t *testing.T) {
	u1 := U{Name: "11"}
	u1.SetName1("22")
	assert.Equal(t, "11", u1.Name)
	u1.SetName("22")
	assert.Equal(t, "22", u1.Name)

	user := User{Name: "abd"}

	vou := reflect.ValueOf(user)
	fld := vou.FieldByName("Name")
	if fld.CanSet() {
		fld.SetString("sss")
	}

	vou = reflect.ValueOf(&user)
	if vou.Kind() == reflect.Ptr {
		vou = vou.Elem()
	}
	fld = vou.FieldByName("Name")
	fld.SetString("sss") // 正确的方法
}
