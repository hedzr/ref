package ref

import (
	"bytes"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

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
	user2 = new(User)
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
		Flags:    []byte("memento"),
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
	Flags    []byte
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

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}

func checkEmployee(employee Employee, user User, t *testing.T, testCase string) {
	if employee.Name != user.Name {
		t.Errorf("%v: Name haven't been copied correctly.", testCase)
	}
	if employee.Nickname == nil || *employee.Nickname != user.Nickname {
		t.Errorf("%v: NickName haven't been copied correctly.", testCase)
	}
	if employee.Birthday == nil && user.Birthday != nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday == nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Age != int64(user.Age) {
		t.Errorf("%v: Age haven't been copied correctly.", testCase)
	}
	if user.FakeAge != nil && employee.FakeAge != int(*user.FakeAge) {
		t.Errorf("%v: FakeAge haven't been copied correctly.", testCase)
	}
	if employee.DoubleAge != user.DoubleAge() {
		t.Errorf("%v: Copy from method doesn't work", testCase)
	}
	if employee.SuperRule != "Super "+user.Role {
		t.Errorf("%v: Copy to method doesn't work", testCase)
	}
	if !reflect.DeepEqual(employee.Notes, user.Notes) {
		t.Errorf("%v: Copy from slice doen't work", testCase)
	}
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
