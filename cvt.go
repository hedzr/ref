package ref

import (
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	strings "strings"
	"unicode"
)

// TryConvert calls reflect.Convert safely, without panic threw.
func TryConvert(v reflect.Value, t reflect.Type) (out reflect.Value, err error) {
	return tryConvert(v, t)
}

func tryConvert(v reflect.Value, t reflect.Type) (out reflect.Value, err error) {
	defer func() {
		if e := recover(); e != nil {
			if e2, ok := e.(error); ok {
				err = e2
			} else {
				err = errors.New("%v", e)
			}
		}
	}()

	out = v.Convert(t)
	return
}

// Captalize returns a copy of the string s with all Unicode letters that begin words
// mapped to their Unicode title case.
//
// BUG(rsc): The rule Title uses for word boundaries does not handle Unicode punctuation properly.
func Captalize(s string) string {
	// Use a closure here to remember state.
	// Hackish but effective. Depends on Map scanning in order and calling
	// the closure once per rune.
	prev := ' '
	return strings.Map(
		func(r rune) rune {
			if isSeparator(prev) {
				prev = r
				return unicode.ToTitle(r)
			}
			prev = r
			return r
		},
		s)
}

// isSeparator reports whether the rune could mark a word boundary.
// TODO: update when package unicode captures more of the properties.
func isSeparator(r rune) bool {
	// ASCII alphanumerics and underscore are not separators
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '-':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}
