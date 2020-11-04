package ref

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	var sb strings.Builder

	DumpEx(user1, func(level int, ownerPath, desc string, v reflect.Value) {
		sb.WriteString(fmt.Sprintf("%v%v", strings.Repeat("  ", level+1), desc))
	}, func(v reflect.Value) {
		sb.WriteString(fmt.Sprintf("\nDump of %v", v.Interface()))
	}, func(v reflect.Value) {
		sb.WriteString("\n")
	})

	t.Log(sb.String())

	sb.Reset()

	//

	Dump(user1, func(level int, ownerPath, desc string, v reflect.Value) {
		sb.WriteString(fmt.Sprintf("%v%v", strings.Repeat("  ", level+1), desc))
	})

	t.Log(sb.String())
}
