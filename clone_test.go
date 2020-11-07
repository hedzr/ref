package ref

import (
	"github.com/hedzr/log"
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/logx/logrus"
	"testing"
	"unsafe"
)

func TestClone(t *testing.T) {
	t.Run("GOB copier: simple clone", testGobSimpleClone)
	t.Run("Default copier: cloneable clone", testDefaultCloneableClone)
	t.Run("Default copier: cloneable clone 2", testDefaultCloneableClone2)
	t.Run("Default copier: simple clone", testDefaultSimpleClone)
	t.Run("Default copier: simple clone - user1", testDefaultSimpleClone2)
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
	// t.Log(u2)

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

func TestCloneMore(t *testing.T) {

	// build.New(build.NewLoggerConfigWith(true, "logex", "debug"))
	l := logrus.New("debug", false, true)
	defer logex.CaptureLog(t).Release()

	l.Infof("123")
	log.Infof("456")

	t.Run("Default copier: copy cov", testDefaultClone_copyCov)
	t.Run("Default copier: copy two struct", testDefaultClone_copyTwoStruct)
	t.Run("Default copier: copy struct", testDefaultClone_copyStruct)
	t.Run("Default copier: copy struct to slice", testDefaultClone_copyFromStructToSlice)

	t.Run("Default copier: on map", testDefaultClone_onMap)

}

func testDefaultClone_copyCov(t *testing.T) {
	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	x0 := X0{}
	x1 := X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:3],
		O: a,
		Q: a,
	}
	x2 := &X2{N: nn[1:3]}

	Clone(&x1, &x2)
}

func testDefaultClone_copyTwoStruct(t *testing.T) {
	user := User{Name: "Real Faked"}
	userTo := User{Name: "Faked", Role: "NN"}
	DefaultCloner.Copy(&user, &userTo)
	t.Log(userTo)
	if userTo.Name != user.Name || userTo.Role != "NN" {
		t.Fatal("wrong")
	}
}

func testDefaultClone_copyStruct(t *testing.T) {

	var fakeAge int32 = 12
	var born int = 7
	var bornU uint = 7
	var sz = "dablo"
	user := User{Name: "Faked", Nickname: "nick"}
	employee := Employee{}

	if err := DefaultCloner.Copy(&user, employee); err == nil {
		t.Errorf("Copy to unaddressable value should get error")
	}

	Clone(&user, &employee, WithIgnoredFieldNames("Shit", "Memo", "Name"))
	// Clone(&employee, &user, "Shit", "Memo", "Name")

	user = User{Name: "Faked", Nickname: "user", Age: 18, FakeAge: &fakeAge,
		Role: "User", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'},
		Retry: 3, Times: 17, RetryU: 23, TimesU: 21, FxReal: 31, FxTime: 37,
		FxTimeU: 13, UxA: 11, UxB: 0, Born: &born, BornU: &bornU,
		Ro: []int{1, 2, 3}, Sptr: &sz, Bool1: true, // Feat: []byte(sz),
	}
	employee = Employee{}
	err := DefaultCloner.Copy(&user, &employee)
	if err != nil {
		t.Errorf("%v", err)
	}
	checkEmployee(employee, user, t, "Copy From Ptr To Ptr")

	employee2 := Employee{}
	Clone(user, &employee2)
	checkEmployee(employee2, user, t, "Copy From Struct To Ptr")

	employee3 := Employee{}
	ptrToUser := &user
	Clone(&ptrToUser, &employee3)
	checkEmployee(employee3, user, t, "Copy From Double Ptr To Ptr")

	employee4 := &Employee{}
	Clone(user, &employee4)
	checkEmployee(*employee4, user, t, "Copy From Ptr To Double Ptr")
}

func testDefaultClone_copyFromStructToSlice(t *testing.T) {
	user := User{Name: "Faked", Age: 18, Role: "User", Notes: []string{"hello world"}}
	employees := []Employee{}

	if err := DefaultCloner.Copy(&user, employees); err != nil && len(employees) != 0 {
		t.Errorf("Copy to unaddressable value should get error")
	}

	if err := DefaultCloner.Copy(&user, &employees); err != nil && len(employees) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(employees[0], user, t, "Copy From Struct To Slice Ptr")
	}

	employees2 := &[]Employee{}
	if err := DefaultCloner.Copy(user, &employees2); err != nil && len(*employees2) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee((*employees2)[0], user, t, "Copy From Struct To Double Slice Ptr")
	}

	employees3 := []*Employee{}
	if err := DefaultCloner.Copy(user, &employees3); err != nil && len(employees3) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*(employees3[0]), user, t, "Copy From Struct To Ptr Slice Ptr")
	}

	employees4 := &[]*Employee{}
	if err := DefaultCloner.Copy(user, &employees4); err != nil && len(*employees4) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*((*employees4)[0]), user, t, "Copy From Struct To Double Ptr Slice Ptr")
	}
}

func testDefaultClone_onMap(t *testing.T) {
	Clone(&uFrom, &uTo)
	t.Log(uTo)
}
