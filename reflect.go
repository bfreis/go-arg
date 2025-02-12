package arg

import (
	"encoding"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/alexflint/go-scalar"
)

var textUnmarshalerType = reflect.TypeOf([]encoding.TextUnmarshaler{}).Elem()

// cardinality tracks how many tokens are expected for a given spec
//   - zero is a boolean, which does to expect any value
//   - one is an ordinary option that will be parsed from a single token
//   - multiple is a slice or map that can accept zero or more tokens
type cardinality int

const (
	zero cardinality = iota
	one
	multiple
	unsupported
)

func (k cardinality) String() string {
	switch k {
	case zero:
		return "zero"
	case one:
		return "one"
	case multiple:
		return "multiple"
	case unsupported:
		return "unsupported"
	default:
		return fmt.Sprintf("unknown(%d)", int(k))
	}
}

// cardinalityOf returns true if the type can be parsed from a string
func cardinalityOf(t reflect.Type) (cardinality, error) {
	if scalar.CanParse(t) {
		if isBoolean(t) {
			return zero, nil
		}
		return one, nil
	}

	// look inside pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// look inside slice and map types
	switch t.Kind() {
	case reflect.Slice:
		if !scalar.CanParse(t.Elem()) {
			return unsupported, fmt.Errorf("cannot parse into %v because %v not supported", t, t.Elem())
		}
		return multiple, nil
	case reflect.Map:
		if !scalar.CanParse(t.Key()) {
			return unsupported, fmt.Errorf("cannot parse into %v because key type %v not supported", t, t.Elem())
		}
		if !scalar.CanParse(t.Elem()) {
			return unsupported, fmt.Errorf("cannot parse into %v because value type %v not supported", t, t.Elem())
		}
		return multiple, nil
	default:
		return unsupported, fmt.Errorf("cannot parse into %v", t)
	}
}

// isBoolean returns true if the type is a boolean or a pointer to a boolean
func isBoolean(t reflect.Type) bool {
	switch {
	case isTextUnmarshaler(t):
		return false
	case t.Kind() == reflect.Bool:
		return true
	case t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Bool:
		return true
	default:
		return false
	}
}

// isTextUnmarshaler returns true if the type or its pointer implements encoding.TextUnmarshaler
func isTextUnmarshaler(t reflect.Type) bool {
	return t.Implements(textUnmarshalerType) || reflect.PtrTo(t).Implements(textUnmarshalerType)
}

// isExported returns true if the struct field name is exported
func isExported(field string) bool {
	r, _ := utf8.DecodeRuneInString(field) // returns RuneError for empty string or invalid UTF8
	return unicode.IsLetter(r) && unicode.IsUpper(r)
}

// isZero returns true if v contains the zero value for its type
func isZero(v reflect.Value) bool {
	t := v.Type()
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Chan || t.Kind() == reflect.Interface {
		return v.IsNil()
	}
	if !t.Comparable() {
		return false
	}
	return v.Interface() == reflect.Zero(t).Interface()
}
