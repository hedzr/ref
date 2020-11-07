package ref

import (
	"fmt"
	"github.com/hedzr/ref/eval"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestDump(t *testing.T) {
	t.Run("Pretty dump: testDumpOnZeroFields", testDumpOnZeroFields)
	t.Run("Pretty dump: testDumpexSimple", testDumpexSimple)
	t.Run("Pretty dump: testDumpOnCircularRef", testDumpOnCircularRef)
}

func testDumpOnZeroFields(t *testing.T) {
	Dump(user2, "user2", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
}

func testDumpexSimple(t *testing.T) {
	var sb strings.Builder

	DumpEx(user1, "usr", func(level int, desc string, v reflect.Value) {
		sb.WriteString(fmt.Sprintf("%v%v\n", strings.Repeat("  ", level+1), desc))
	}, func(v reflect.Value) {
		sb.WriteString(fmt.Sprintf("\nDump of %v\n", v.Interface()))
	}, func(v reflect.Value) {
		sb.WriteString("\n")
	})

	t.Log(sb.String())

	//

	//sb.Reset()
	//
	//Dump(user1, "user1", func(level int, ownerPath, desc string, v reflect.Value) {
	//	sb.WriteString(fmt.Sprintf("%v%v", strings.Repeat("  ", level+1), desc))
	//})
	//
	//t.Log(sb.String())
}

func testDumpOnCircularRef(t *testing.T) {
	// Circular linked lists a -> b -> a and c -> c.
	type link struct {
		Value string
		Tail  *link
	}
	a, b, c := &link{Value: "a"}, &link{Value: "b"}, &link{Value: "c"}
	a.Tail, b.Tail, c.Tail = b, a, c
	fmt.Println(Equal(a, a)) // "true"
	fmt.Println(Equal(b, b)) // "true"
	fmt.Println(Equal(c, c)) // "true"
	fmt.Println(Equal(a, b)) // "false"
	fmt.Println(Equal(a, c)) // "false"

	Dump(a, "a", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	Dump(b, "b", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	Dump(c, "c", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
}

// NOTE: we can't use !+..!- comments to excerpt these tests
// into the book because it defeats the Example mechanism,
// which requires the // Output comment to be at the end
// of the function.

func Example_expr() {
	e, _ := eval.Parse("sqrt(A / pi)")
	Dump(e, "e", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display e (eval.call):
	// e.fn = "sqrt"
	// e.args[0].type = eval.binary
	// e.args[0].value.op = 47
	// e.args[0].value.x.type = eval.Var
	// e.args[0].value.x.value = "A"
	// e.args[0].value.y.type = eval.Var
	// e.args[0].value.y.value = "pi"
}

func Example_slice() {
	Dump([]*int{new(int), nil}, "slice", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display slice ([]*int):
	// (*slice[0]) = 0
	// slice[1] = nil
}

func Example_nilInterface() {
	var w io.Writer
	Dump(w, "w", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display w (<nil>):
	// w = invalid
}

func Example_ptrToInterface() {
	var w io.Writer
	Dump(&w, "&w", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display &w (*io.Writer):
	// (*&w) = nil
}

func Example_struct() {
	Dump(struct{ x interface{} }{3}, "x", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display x (struct { x interface {} }):
	// x.x.type = int
	// x.x.value = 3
}

func Example_interface() {
	var i interface{} = 3
	Dump(i, "i", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display i (int):
	// i = 3
}

func Example_ptrToInterface2() {
	var i interface{} = 3
	Dump(&i, "&i", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display &i (*interface {}):
	// (*&i).type = int
	// (*&i).value = 3
}

func Example_array() {
	Dump([1]interface{}{3}, "x", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display x ([1]interface {}):
	// x[0].type = int
	// x[0].value = 3
}

func Example_movie() {
	//!+movie
	type Movie struct {
		Title, Subtitle string
		Year            int
		Color           bool
		Actor           map[string]string
		Oscars          []string
		Sequel          *string
	}
	//!-movie
	//!+strangelove
	strangelove := Movie{
		Title:    "Dr. Strangelove",
		Subtitle: "How I Learned to Stop Worrying and Love the Bomb",
		Year:     1964,
		Color:    false,
		Actor: map[string]string{
			"Dr. Strangelove":            "Peter Sellers",
			"Grp. Capt. Lionel Mandrake": "Peter Sellers",
			"Pres. Merkin Muffley":       "Peter Sellers",
			"Gen. Buck Turgidson":        "George C. Scott",
			"Brig. Gen. Jack D. Ripper":  "Sterling Hayden",
			`Maj. T.J. "King" Kong`:      "Slim Pickens",
		},

		Oscars: []string{
			"Best Actor (Nomin.)",
			"Best Adapted Screenplay (Nomin.)",
			"Best Director (Nomin.)",
			"Best Picture (Nomin.)",
		},
	}
	//!-strangelove
	Dump(strangelove, "strangelove", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})

	// We don't use an Output: comment since displaying
	// a map is nondeterministic.
	/*
		//!+output
		Display strangelove (display.Movie):
		strangelove.Title = "Dr. Strangelove"
		strangelove.Subtitle = "How I Learned to Stop Worrying and Love the Bomb"
		strangelove.Year = 1964
		strangelove.Color = false
		strangelove.Actor["Gen. Buck Turgidson"] = "George C. Scott"
		strangelove.Actor["Brig. Gen. Jack D. Ripper"] = "Sterling Hayden"
		strangelove.Actor["Maj. T.J. \"King\" Kong"] = "Slim Pickens"
		strangelove.Actor["Dr. Strangelove"] = "Peter Sellers"
		strangelove.Actor["Grp. Capt. Lionel Mandrake"] = "Peter Sellers"
		strangelove.Actor["Pres. Merkin Muffley"] = "Peter Sellers"
		strangelove.Oscars[0] = "Best Actor (Nomin.)"
		strangelove.Oscars[1] = "Best Adapted Screenplay (Nomin.)"
		strangelove.Oscars[2] = "Best Director (Nomin.)"
		strangelove.Oscars[3] = "Best Picture (Nomin.)"
		strangelove.Sequel = nil
		//!-output
	*/
}

// This test ensures that the program terminates without crashing.
func TestDumpMore(t *testing.T) {
	// Some other values (YMMV)
	Dump(os.Stderr, "os.Stderr", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display os.Stderr (*os.File):
	// (*(*os.Stderr).file).fd = 2
	// (*(*os.Stderr).file).name = "/dev/stderr"
	// (*(*os.Stderr).file).nepipe = 0

	var w io.Writer = os.Stderr
	Dump(&w, "&w", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display &w (*io.Writer):
	// (*&w).type = *os.File
	// (*(*(*&w).value).file).fd = 2
	// (*(*(*&w).value).file).name = "/dev/stderr"
	// (*(*(*&w).value).file).nepipe = 0

	var locker sync.Locker = new(sync.Mutex)
	Dump(&locker, "(&locker)", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display (&locker) (*sync.Locker):
	// (*(&locker)).type = *sync.Mutex
	// (*(*(&locker)).value).state = 0
	// (*(*(&locker)).value).sema = 0

	Dump(locker, "locker", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display locker (*sync.Mutex):
	// (*locker).state = 0
	// (*locker).sema = 0
	// (*(&locker)) = nil

	locker = nil
	Dump(&locker, "(&locker)", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display (&locker) (*sync.Locker):
	// (*(&locker)) = nil

	ips, _ := net.LookupHost("msn.com")
	Dump(ips, "ips", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display ips ([]string):
	// ips[0] = "173.194.68.141"
	// ips[1] = "2607:f8b0:400d:c06::8d"

	// Even metarecursion!  (YMMV)
	Dump(reflect.ValueOf(os.Stderr), "rV", func(level int, desc string, v reflect.Value) {
		fmt.Println(strings.Repeat("  ", level+1), desc)
	})
	// Output:
	// Display rV (reflect.Value):
	// (*rV.typ).size = 8
	// (*rV.typ).ptrdata = 8
	// (*rV.typ).hash = 871609668
	// (*rV.typ)._ = 0
	// ...

	// a pointer that points to itself
	type P *P
	var p P
	p = &p
	if false {
		Dump(p, "p", func(level int, desc string, v reflect.Value) {
			fmt.Println(strings.Repeat("  ", level+1), desc)
		})
		// Output:
		// Display p (display.P):
		// ...stuck, no output...
	}

	// a map that contains itself
	type M map[string]M
	m := make(M)
	m[""] = m
	if false {
		Dump(m, "m", func(level int, desc string, v reflect.Value) {
			fmt.Println(strings.Repeat("  ", level+1), desc)
		})
		// Output:
		// Display m (display.M):
		// ...stuck, no output...
	}

	// a slice that contains itself
	type S []S
	s := make(S, 1)
	s[0] = s
	if false {
		Dump(s, "s", func(level int, desc string, v reflect.Value) {
			fmt.Println(strings.Repeat("  ", level+1), desc)
		})
		// Output:
		// Display s (display.S):
		// ...stuck, no output...
	}

	// a linked list that eats its own tail
	type Cycle struct {
		Value int
		Tail  *Cycle
	}
	var c Cycle
	c = Cycle{42, &c}
	if false {
		Dump(c, "c", func(level int, desc string, v reflect.Value) {
			fmt.Println(strings.Repeat("  ", level+1), desc)
		})
		// Output:
		// Display c (display.Cycle):
		// c.Value = 42
		// (*c.Tail).Value = 42
		// (*(*c.Tail).Tail).Value = 42
		// ...ad infinitum...
	}
}
