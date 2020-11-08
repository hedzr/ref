package ref

import (
	"github.com/hedzr/log"
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/logx/logrus"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	"testing"
	"time"
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
	t.Run("Default copier: copy embed", testDefaultClone_copyEmbedded)
	t.Run("Default copier: copy fields with same name but different types", testDefaultClone_copyFieldsWithSameNameButDifferentTypes)

	t.Run("Default copier: test scanner", testDefaultClone_testScanner)
	t.Run("Default copier: copy between primitive types", testDefaultClone_copyBetweenPrimitiveTypes)
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

func testDefaultClone_copyEmbedded(t *testing.T) {
	type Base struct {
		BaseField1 int
		BaseField2 int
	}

	type Embed struct {
		EmbedField1 int
		EmbedField2 int
		Base
	}

	base := Base{}
	embedded := Embed{}
	embedded.BaseField1 = 1
	embedded.BaseField2 = 2
	embedded.EmbedField1 = 3
	embedded.EmbedField2 = 4

	Clone(&embedded, &base)

	if base.BaseField1 != 1 {
		t.Error("Embedded fields not copied")
	}

	if err := DefaultCloner.Copy(&embedded, &base); err != nil {
		t.Error(err)
	}
}

type structSameName1 struct {
	A string
	B int64
	C time.Time
}

type structSameName2 struct {
	A string
	B time.Time
	C int64
}

func testDefaultClone_copyFieldsWithSameNameButDifferentTypes(t *testing.T) {
	obj1 := structSameName1{A: "123", B: 2, C: time.Now()}
	obj2 := &structSameName2{}
	err := DefaultCloner.Copy(&obj1, obj2)
	if err != nil {
		t.Logf("For the diff types some error should be raised but you can omit them safely: %v", err)
	}

	if obj2.A != obj1.A {
		t.Errorf("Field A should be copied")
	}

	err = DefaultCloner.Copy(&obj1, obj2)
	if err != nil {
		// t.Error(err)
		t.Logf("For the diff types some error should be raised but you can omit them safely: %v", err)
	}
}

type ScannerValue struct {
	V int
}

func (s *ScannerValue) Scan(src interface{}) error {
	return errors.New("I failed")
}

type ScannerStruct struct {
	V *ScannerValue
}

type ScannerStructTo struct {
	V *ScannerValue
}

func testDefaultClone_testScanner(t *testing.T) {
	s := &ScannerStruct{
		V: &ScannerValue{
			V: 12,
		},
	}

	s2 := &ScannerStructTo{}

	err := DefaultCloner.Copy(s, s2)
	if err != nil {
		t.Error("Should not raise error")
	}

	if s.V.V != s2.V.V {
		t.Errorf("Field V should be copied")
	}
}

func testDefaultClone_copyBetweenPrimitiveTypes(t *testing.T) {
	var aa = "dsajkld"
	var b int

	// cmdr.Clone(b, aa)

	Clone(&aa, b)

	Clone(&aa, &b)

	Clone(nil, &b)
	var b1 *int = &b
	Clone(nil, &b1)

	var c, d bool
	Clone(&c, &d)

	var e, f int
	Clone(&e, &f)
	var e1, f1 int8
	f1 = 1
	Clone(&f1, &e1)
	if e1 != 1 {
		t.Failed()
	}
	var e2, f2 int16
	Clone(&e2, &f2)
	var e3, f3 int32
	Clone(&e3, &f3)
	var e4, f4 int64
	e4 = 9
	Clone(&e4, &f4)
	if f4 != 9 {
		t.Failed()
	}

	var g, h string
	Clone(&g, &h)
}

func TestCloneMap(t *testing.T) {

	t.Run("Map: basics", testMap_basics)
	t.Run("Map: new map 1", testMap_newMap_1)
	t.Run("Map: new map 2", testMap_newMap_2)
	t.Run("Map: clone", testMap_clone)

}

func testMap_newMap_1(t *testing.T) {
	var key = "key1"
	var value = 123

	var keyType = reflect.TypeOf(key)
	var valueType = reflect.TypeOf(value)
	var aMapType = reflect.MapOf(keyType, valueType)
	aMap := reflect.MakeMapWithSize(aMapType, 0)
	aMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	t.Logf("%T:  %v\n", aMap.Interface(), aMap.Interface())
}

func testMap_newMap_2(t *testing.T) {
	key := 1
	value := "abc"

	mapType := reflect.MapOf(reflect.TypeOf(key), reflect.TypeOf(value))

	mapValue := reflect.MakeMap(mapType)
	mapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	mapValue.SetMapIndex(reflect.ValueOf(2), reflect.ValueOf("def"))
	mapValue.SetMapIndex(reflect.ValueOf(3), reflect.ValueOf("gh"))

	keys := mapValue.MapKeys()
	for _, k := range keys {
		ck := k.Convert(mapValue.Type().Key())
		cv := mapValue.MapIndex(ck)
		t.Logf("key: %v,  value: %v", ck, cv)
	}
}

func testMap_basics(t *testing.T) {
	mi := map[string]string{
		"a": "this is a",
		"b": "this is b",
	}

	var input interface{}
	input = mi
	m := reflect.ValueOf(input)
	if m.Kind() == reflect.Map {

		var key = reflect.ValueOf("a")
		val := m.MapIndex(key)
		t.Logf("map[%q] = %q", key.String(), val.String())

		var it *reflect.MapIter
		it = m.MapRange()
		t.Log("iterating...")
		for it.Next() {
			t.Logf("  m[%q] = %q", it.Key(), it.Value())
		}

		newInstance := reflect.MakeMap(m.Type())
		keys := m.MapKeys()
		for _, k := range keys {
			key := k.Convert(newInstance.Type().Key()) //.Convert(m.Type().Key())
			value := m.MapIndex(key)
			newInstance.SetMapIndex(key, value)
		}
		t.Logf("newInstance = %v", newInstance)
	}
}

func testMap_clone(t *testing.T) {
	var m1 = map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "text",
	}
	var m2 map[string]interface{}

	Clone(m1, &m2)
	t.Log(m2)
}
