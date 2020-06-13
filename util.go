package transformation

import (
	"fmt"
	"reflect"
)

// findStructField looks for a field in the given struct.
// The field being looked for should be a pointer to the actual struct field.
// If found, the field info will be returned. Otherwise, nil will be returned.
func findStructField(structValue reflect.Value, fieldValue reflect.Value) *reflect.StructField {
	ptr := fieldValue.Pointer()
	for i := structValue.NumField() - 1; i >= 0; i-- {
		sf := structValue.Type().Field(i)
		if ptr == structValue.Field(i).UnsafeAddr() {
			// do additional type comparison because it's possible that the address of
			// an embedded struct is the same as the first field of the embedded struct
			if sf.Type == fieldValue.Elem().Type() {
				return &sf
			}
		}
		if sf.Anonymous {
			// delve into anonymous struct to look for the field
			fi := structValue.Field(i)
			if sf.Type.Kind() == reflect.Ptr {
				fi = fi.Elem()
			}
			if fi.Kind() == reflect.Struct {
				if f := findStructField(fi, fieldValue); f != nil {
					return f
				}
			}
		}
	}
	return nil
}

func CopyValue(src interface{}, dest interface{}) error {
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("destination expected to be a pointer but got %T", dest)
	}

	srcRef := reflect.ValueOf(src)
	vp := reflect.ValueOf(dest)
	if vp.Elem().Kind() != reflect.Ptr {
		srcRef = reflect.Indirect(srcRef)
	}

	vp.Elem().Set(srcRef)

	return nil
}

func MustCopyValue(src interface{}, dest interface{}) {
	if err := CopyValue(src, dest); err != nil {
		panic(err)
	}
}

// Indirect returns the value that the given interface or pointer references to.
// If the value implements driver.Valuer, it will deal with the value returned by
// the Value() method instead. A boolean value is also returned to indicate if
// the value is nil or not (only applicable to interface, pointer, map, and slice).
// If the value is neither an interface nor a pointer, it will be returned back.
func Indirect(value interface{}) (interface{}, bool) {
	rv := reflect.ValueOf(value)
	kind := rv.Kind()
	switch kind {
	case reflect.Invalid:
		return nil, true
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil, true
		}
		return Indirect(rv.Elem().Interface())
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		if rv.IsNil() {
			return nil, true
		}
	}

	return value, false
}
