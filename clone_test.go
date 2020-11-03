package ref

import (
	"bytes"
	"github.com/hedzr/assert"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestClone(t *testing.T) {
	t.Run("reflect set struct field", testSetStructField)

	t.Run("GOB copier: simple clone", testGobSimpleClone)
	t.Run("Default copier: cloneable clone", testDefaultCloneableClone)
	t.Run("Default copier: cloneable clone 2", testDefaultCloneableClone2)
	t.Run("Default copier: simple clone", testDefaultSimpleClone)
	t.Run("Default copier: simple clone - user1", testDefaultSimpleClone2)

	t.Run("Default copier: map", testDefaultCloneOnMap)

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

func testGobSimpleClone(t *testing.T) {
	err := LazyGobCopier.Copy(&wilson, &nikita)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log(nikita)
}

func testDefaultCloneableClone(t *testing.T) {
	err := DefaultCloner.Copy(&uFrom, &uTo)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log(nikita)
}

func testDefaultCloneableClone2(t *testing.T) {
	Clone(&uFrom, &uTo)
	t.Log(uTo)
}

func testDefaultSimpleClone(t *testing.T) {
	err := DefaultCloner.Copy(&wilson, &nikita)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log(nikita)
}

func testDefaultSimpleClone2(t *testing.T) {
	var u1 *User = new(User)
	var u2 User

	err := DefaultCloner.Copy(user1, u1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log(u1)

	err = DefaultCloner.Copy(user1, u2)
	if err == nil {
		t.Fatal("expecting error return but missed it: 'target cannot be set: field \"Name\" (value: Buggy Forman)'")
	}
	//t.Log(u2)

	err = DefaultCloner.Copy(&user1, u1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log(u1)

	err = DefaultCloner.Copy(&user1, &u2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log(u2)
}

func testDefaultCloneOnMap(t *testing.T) {
	Clone(&uFrom, &uTo)
	t.Log(uTo)
}

var (
	wilson          = Cat{7, "Wilson", []string{"Tom", "Tabata", "Willie"}}
	nikita          = Cat{}
	now             = time.Now()
	older, olderErr = time.Parse("", "1999-12-31")
	tomAgeI32       = int32(13)
	tomBornI        = int(15)
	tomBornU        = uint(17)
	tomSleep        = "tom is in sleep."
	uFrom           = U{
		Name:     "123",
		Birthday: &now,
		Nickname: "456",
	}
	uTo   = &U{}
	user1 = User{
		Name:     "Buggy Forman",
		Birthday: &older,
		Nickname: "bfboy",
		Role:     "normal",
		Age:      21,
		Retry:    -5,
		Times:    -2,
		RetryU:   121,
		TimesU:   7013,
		FxReal:   144,
		FxTime:   169,
		FxTimeU:  1999703,
		UxA:      8,
		UxB:      6,
		FakeAge:  &tomAgeI32,
		Notes:    []string{"fds", "kjrtl", "re34"},
		flags:    []byte("memento"),
		Born:     &tomBornI,
		BornU:    &tomBornU,
		Ro:       []int{56, 78, 90, 12},
		F11:      3.14,
		F12:      3.14159,
		C11:      3.14 + 9i,
		C12:      3.1415926535 + 111i,
		Sptr:     &tomSleep,
		Bool1:    false,
		Bool2:    true,
	}
)

type Cat struct {
	Age     int
	name    string
	friends []string
}

type U struct {
	Name     string
	Birthday *time.Time
	Nickname string
}

func (u U) SetName1(n string) { u.Name = n }
func (u *U) SetName(n string) { u.Name = n }
func (u U) Clone() interface{} {
	return &U{
		Name:     u.Name,
		Birthday: u.Birthday,
		Nickname: u.Nickname,
	}
}

type User struct {
	Name     string
	Birthday *time.Time
	Nickname string
	Role     string
	Age      int32
	Retry    int8
	Times    int16
	RetryU   uint8
	TimesU   uint16
	FxReal   uint32
	FxTime   int64
	FxTimeU  uint64
	UxA      uint
	UxB      int
	FakeAge  *int32
	Notes    []string
	flags    []byte
	Born     *int
	BornU    *uint
	Ro       []int
	F11      float32
	F12      float64
	C11      complex64
	C12      complex128
	Sptr     *string
	Bool1    bool
	Bool2    bool
	// Feat     []byte
}

func (user User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	Name      string
	Birthday  *time.Time
	F11       float32
	F12       float64
	C11       complex64
	C12       complex128
	Feat      []byte
	Sptr      *string
	Nickname  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	RetryU    uint8
	TimesU    uint16
	FxReal    uint32
	FxTime    int64
	FxTimeU   uint64
	UxA       uint
	UxB       int
	Retry     int8
	Times     int16
	Born      *int
	BornU     *uint
	flags     []byte
	Bool1     bool
	Bool2     bool
	Ro        []int
}

type X0 struct{}

type X1 struct {
	A uintptr
	B map[string]interface{}
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J interface{}
	K *X0
	L unsafe.Pointer
	M unsafe.Pointer
	N []int
	O [2]string
	P [2]string
	Q [2]string
}

type X2 struct {
	A uintptr
	B map[string]interface{}
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J interface{}
	K *X0
	L unsafe.Pointer
	M unsafe.Pointer
	N []int
	O [2]string
	P [2]string
	Q [3]string
}
