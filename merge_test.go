package ref

import (
	"github.com/hedzr/assert"
	"testing"
)

func TestMergeMap(t *testing.T) {
	defer initLogger(t)()

	t.Run("struct -> struct", testMergeStructStruct)
	t.Run("map -> struct", testMergeMapStruct)
	t.Run("basics", testMergeMapBasics)
	t.Run("primitive types", testMergePrimitiveTypes)
}

func testMergePrimitiveTypes(t *testing.T) {
	var err error

	var vb bool
	var vi int
	var vu uint
	var vs string

	if err = NewMerger(true).MergeTo(&vb); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("vb = %v", vb)

	if err = NewMerger(-1).MergeTo(&vi); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("vi = %v", vi)

	if err = NewMerger(1).MergeTo(&vu); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("vu = %v", vu)

	if err = NewMerger("str").MergeTo(&vs); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("vs = %v", vs)

	if err = NewMerger(89).MergeTo(&vs); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("vs = %v", vs)
	assert.Equal(t, "Y", vs)

	if err = NewMerger(65).MergeTo(&vs); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("vs = %v", vs)
	assert.Equal(t, "A", vs)

}

func testMergeStructStruct(t *testing.T) {
	type sss struct {
		A int64
		B bool
		C string
	}

	var err error
	var s1 = sss{1, true, "world"}
	mm := NewMerger(s1)

	var s2 sss
	if err = mm.MergeTo(&s2); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("sss / s2 = %v", s2)
}

func testMergeMapStruct(t *testing.T) {
	var m1 = map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "text",
	}
	type sss struct {
		A int64
		B bool
		C string
	}

	var err error
	mm := NewMerger(m1)

	var s sss
	if err = mm.MergeTo(&s); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("sss / s = %v", s)
	assert.Equal(t, int64(1), s.A)
	assert.Equal(t, true, s.B)
	assert.Equal(t, "text", s.C)
}

func testMergeMapBasics(t *testing.T) {

	m1 := map[string]interface{}{
		"a": 1,
		"b": int64(2),
		"c": true,
		"d": "ff",
		"e": []string{"aaa", "bbb"},
		"f": &user1,
		"g": map[string]interface{}{
			"a": 1,
			"b": int64(7),
			"c": true,
			"d": "ff",
			"e": []string{"aa", "bb"},
			"f": &user1,
		},
	}

	var err error
	mm := NewMerger(m1)

	mm.Reset()
	var m2 map[string]interface{}
	if err = mm.MergeTo(&m2); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("m2 = %v", m2)

	mm.Reset()
	var m3 map[int]interface{}
	if err = mm.MergeTo(&m3); err == nil {
		t.Fatalf("should be panic for non-assignable type")
	} else {
		t.Logf("the expected error is: %v", err)
	}

	var m4 = map[string]interface{}{
		"g": map[string]interface{}{
			"b": int32(1),
			"e": []string{"bb", "cc"},
			"f": &user2,
		},
	}
	mm.Reset()
	if err = mm.MergeTo(&m4); err != nil {
		t.Fatalf("merge map error: %v", err)
	}
	t.Logf("m4 = %v", m4)
	//Dump(m4, "m4", func(level int, desc string, v reflect.Value) {
	//	t.Logf("%v%v", strings.Repeat("  ", level+1), desc)
	//})
	assert.Equal(t, 1, m4["a"])
	assert.Equal(t, int64(2), m4["b"])
	assert.Equal(t, true, m4["c"])
	assert.Equal(t, "ff", m4["d"])
	assert.Equal(t, []string{"aa", "bb", "cc"}, m4["g"].(map[string]interface{})["e"])
	assert.Equal(t, user1, m4["f"])
}
