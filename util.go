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

func copyValue(src interface{}, dest interface{}) error {
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("destination expected to be a pointer but got %T", dest)
	}

	src, isNil := indirect(src)
	if isNil {
		return nil
	}

	srcValue := reflect.ValueOf(src)
	destValue := reflect.ValueOf(dest)
	switch srcValue.Kind() {
	case reflect.Slice, reflect.Array:
		return copySlice(src, dest)
	case reflect.Map:
		panic("not implemented")
	default:
		if destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Ptr && srcValue.Kind() != reflect.Ptr {
			srcValue = makePtr(srcValue)
		}

		if destValue.Elem().Kind() != reflect.Ptr {
			srcValue = reflect.Indirect(srcValue)
		}

		destValue.Elem().Set(srcValue)

		return nil
	}
}

func mustCopyValue(src interface{}, dest interface{}) {
	if err := copyValue(src, dest); err != nil {
		panic(err)
	}
}

// indirect returns the value that the given interface or pointer references to.
// If the value implements driver.Valuer, it will deal with the value returned by
// the Value() method instead. A boolean value is also returned to indicate if
// the value is nil or not (only applicable to interface, pointer, map, and slice).
// If the value is neither an interface nor a pointer, it will be returned back.
func indirect(value interface{}) (interface{}, bool) {
	rv := reflect.ValueOf(value)
	kind := rv.Kind()
	switch kind {
	case reflect.Invalid:
		return nil, true
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil, true
		}
		return indirect(rv.Elem().Interface())
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		if rv.IsNil() {
			return nil, true
		}
	}

	return value, false
}

func makePtr(v reflect.Value) reflect.Value {
	nv := reflect.New(v.Type())
	nv.Elem().Set(v)

	return nv
}

func toConcreteSlice(from []interface{}) interface{} {
	fromValue := reflect.ValueOf(from)
	t := reflect.TypeOf(from).Elem()
	if sliceHasSameType(from) {
		t = fromValue.Index(0).Elem().Type()
	}

	st := reflect.SliceOf(t)
	sl := reflect.MakeSlice(st, fromValue.Len(), fromValue.Len())
	for i := 0; i < fromValue.Len(); i++ {
		sl.Index(i).Set(fromValue.Index(i).Elem())
	}

	return sl.Interface()
}

func toConcreteMap(from map[interface{}]interface{}) interface{} {
	hasSameValueType := mapHasSameValueType(from)
	hasSameKeyType := mapHasSameKeyType(from)
	if !hasSameValueType && !hasSameKeyType {
		return from
	}

	fromValue := reflect.ValueOf(from)
	k := fromValue.MapKeys()[0]
	kt := fromValue.MapKeys()[0].Type()
	if hasSameKeyType {
		kt = fromValue.MapKeys()[0].Elem().Type()
	}

	vt := fromValue.MapIndex(k).Type()
	if hasSameValueType {
		vt = fromValue.MapIndex(k).Elem().Type()
	}

	mt := reflect.MapOf(kt, vt)
	m := reflect.MakeMap(mt)
	iter := fromValue.MapRange()
	for iter.Next() {
		m.SetMapIndex(iter.Key().Elem(), iter.Value().Elem())
	}

	return m.Interface()
}

func mapHasSameValueType(i map[interface{}]interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Len() == 0 {
		return false
	}

	if v.Len() == 1 {
		return true
	}

	iter := v.MapRange()
	iter.Next()
	t := iter.Value().Elem().Type()
	for iter.Next() {
		if t != iter.Value().Elem().Type() {
			return false
		}
	}

	return true
}

func mapHasSameKeyType(i map[interface{}]interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Len() == 0 {
		return false
	}

	if v.Len() == 1 {
		return true
	}

	iter := v.MapRange()
	iter.Next()
	t := iter.Key().Elem().Type()
	for iter.Next() {
		if t != iter.Key().Elem().Type() {
			return false
		}
	}

	return true
}

func sliceHasSameType(i []interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Len() == 0 {
		return false
	}

	if v.Len() == 1 {
		return true
	}

	t := v.Index(0).Elem().Type()
	for i := 1; i < v.Len(); i++ {
		if t != v.Index(i).Elem().Type() {
			return false
		}
	}

	return true
}

func copySlice(src interface{}, dest interface{}) error {
	srcValue := reflect.ValueOf(src)
	destValue := reflect.ValueOf(dest)
	sl := makeSliceFrom(destValue, srcValue.Len(), srcValue.Len())
	for i := 0; i < srcValue.Len(); i++ {
		el := reflect.New(sl.Index(i).Type())
		if err := copyValue(srcValue.Index(i).Interface(), el.Interface()); err != nil {
			return err
		}
		sl.Index(i).Set(el.Elem())
	}

	assign(destValue, sl)

	return nil
}

func assign(to, from reflect.Value) {
	if to.Kind() != reflect.Ptr {
		panic(fmt.Errorf("expected %s but got %s", reflect.Ptr, to.Kind()))
	}

	if to.Elem().Kind() == reflect.Ptr && from.Kind() != reflect.Ptr {
		from = makePtr(from)
	}

	if from.Kind() == reflect.Ptr && to.Elem().Kind() != reflect.Ptr {
		from = from.Elem()
	}

	to.Elem().Set(from)
}

func makeSliceFrom(from reflect.Value, len, cap int) reflect.Value {
	st := from.Type()
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}

	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}

	return reflect.MakeSlice(st, len, cap)
}
