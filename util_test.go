package transformation_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/vcraescu/go-transformation"
	"reflect"
	"testing"
	"time"
)

func TestCopyValue(t *testing.T) {
	from := "test"
	var to *string

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.NotNil(t, to)
		assert.Equal(t, from, *to)
	}
}

func TestCopyValueSlice(t *testing.T) {
	from := []string{"1 ", "2", "3"}
	var to []*string

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		for i, v := range to {
			if assert.NotNil(t, v) {
				assert.Equal(t, from[i], *v)
			}
		}
	}
}

func TestCopyValueSlice2(t *testing.T) {
	from := []string{"1 ", "2", "3"}
	var to []string

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		for i, v := range to {
			assert.Equal(t, from[i], v)
		}
	}
}

func TestCopyValueSlice3(t *testing.T) {
	from := []time.Time{time.Now(), time.Now(), time.Now()}
	var to []*time.Time

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		for i, v := range to {
			assert.Equal(t, from[i].Unix(), v.Unix())
		}
	}
}

func TestCopyValueSlice4(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now()
	from := []*time.Time{&t1, &t2}
	var to []time.Time

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		for i, v := range to {
			assert.Equal(t, from[i].Unix(), v.Unix())
		}
	}
}

func TestCopyValueSlice5(t *testing.T) {
	var from []*time.Time
	var to []time.Time

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
	}
}

func TestCopyValueSlice6(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now()
	var from *[]*time.Time
	values := []*time.Time{&t1, &t2}
	from = &values
	var to []time.Time

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(*from))
	}
}

func TestCopyValueSlice7(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now()
	var from *[]*time.Time
	values := []*time.Time{&t1, &t2}
	from = &values
	var to *[]time.Time

	err := transformation.CopyValue(from, &to)
	if assert.NoError(t, err) {
		assert.NotNil(t, to)
		assert.Len(t, *to, len(*from))
	}
}

func TestMakePtr(t *testing.T) {
	s := "test"
	sp := transformation.MakePtr(reflect.ValueOf(s))
	if assert.Equal(t, reflect.Ptr, sp.Kind()) {
		assert.Equal(t, "test", sp.Elem().String())
	}

	var a *int
	b := 1
	a = &b
	ap := transformation.MakePtr(reflect.ValueOf(a))
	if assert.Equal(t, reflect.Ptr, ap.Kind()) {
		assert.Equal(t, int64(b), ap.Elem().Elem().Int())
	}
}

func TestMakeSliceFrom(t *testing.T) {
	var a []string
	va := transformation.MakeSliceFrom(reflect.ValueOf(a), 0, 0)
	_, ok := va.Interface().([]string)
	assert.True(t, ok)

	var b *[]string
	vb := transformation.MakeSliceFrom(reflect.ValueOf(b), 0, 0)
	_, ok = vb.Interface().([]string)
	assert.True(t, ok)

	var c *[]string
	vc := transformation.MakeSliceFrom(reflect.ValueOf(&c), 0, 0)
	_, ok = vc.Interface().([]string)
	assert.True(t, ok)
}

func TestAssign(t *testing.T) {
	a1 := []string{"1", "2"}
	var b1 []string
	transformation.Assign(reflect.ValueOf(&b1), reflect.ValueOf(a1))
	assert.Len(t, b1, len(a1))

	var a2 *[]string
	a2 = &[]string{"1", "2"}
	var b2 *[]string
	transformation.Assign(reflect.ValueOf(&b2), reflect.ValueOf(a2))
	assert.Len(t, *b2, len(*a2))

	var a3 []string
	a3 = []string{"1", "2"}
	var b3 *[]string
	transformation.Assign(reflect.ValueOf(&b3), reflect.ValueOf(a3))
	assert.Len(t, *b3, len(a3))

	var a4 *[]string
	a4 = &[]string{"1", "2"}
	var b4 []string
	transformation.Assign(reflect.ValueOf(&b4), reflect.ValueOf(a4))
	assert.Len(t, b4, len(*a4))

	a5 := 1
	b5 := 2
	transformation.Assign(reflect.ValueOf(&b5), reflect.ValueOf(a5))
	assert.Equal(t, a5, b5)
}

func TestCopyConcretSlice(t *testing.T) {
	a1 := []interface{}{"1", "2", "3"}
	r := transformation.ToConcreteSlice(a1)
	b1, ok := r.([]string)
	if assert.True(t, ok) {
		if assert.Len(t, b1, len(a1)) {
			assert.Equal(t, "1", b1[0])
			assert.Equal(t, "2", b1[1])
			assert.Equal(t, "3", b1[2])
		}
	}

	v1 := "1"
	a2 := []interface{}{&v1}
	r = transformation.ToConcreteSlice(a2)
	b2, ok := r.([]*string)
	if assert.True(t, ok) {
		if assert.Len(t, b2, len(a2)) {
			assert.Equal(t, "1", *b2[0])
		}
	}

	a3 := []interface{}{&v1, "2"}
	r = transformation.ToConcreteSlice(a3)
	b3, ok := r.([]interface{})
	if assert.True(t, ok) {
		assert.Len(t, b3, len(a3))
	}
}

func TestSliceHasSameType(t *testing.T) {
	a := 1

	tests := []struct {
		value    []interface{}
		expected bool
	}{
		{[]interface{}{}, false},
		{[]interface{}{1}, true},
		{[]interface{}{1, 2}, true},
		{[]interface{}{1, 2, "3"}, false},
		{[]interface{}{1, 2, &a}, false},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, transformation.SliceHasSameType(test.value))
	}
}

func TestMapHasSameValueType(t *testing.T) {
	v1 := "test"
	tests := []struct {
		value    map[interface{}]interface{}
		expected bool
	}{
		{map[interface{}]interface{}{1: "test", 2: "test"}, true},
		{map[interface{}]interface{}{1: "test", 2: 3}, false},
		{map[interface{}]interface{}{1: &v1, 2: "test"}, false},
		{map[interface{}]interface{}{1: time.Now(), 2: time.Now()}, true},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, transformation.MapHasSameValueType(test.value))
	}
}

func TestMapHasSameKeyType(t *testing.T) {
	tests := []struct {
		value    map[interface{}]interface{}
		expected bool
	}{
		{map[interface{}]interface{}{1: "test", 2: "test"}, true},
		{map[interface{}]interface{}{"t": "test", 2: 3}, false},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, transformation.MapHasSameKeyType(test.value))
	}
}

func TestToConcreteMap(t *testing.T) {
	tests := []struct {
		value  map[interface{}]interface{}
		typ    reflect.Type
		length int
	}{
		{map[interface{}]interface{}{1: "test", 2: "test"}, reflect.TypeOf((map[int]string)(nil)), 2},
		{map[interface{}]interface{}{1: "test", 2: 2}, reflect.TypeOf((map[int]interface{})(nil)), 2},
		{map[interface{}]interface{}{"test": 2, 2: 2}, reflect.TypeOf((map[interface{}]int)(nil)), 2},
	}
	for _, test := range tests {
		to := transformation.ToConcreteMap(test.value)
		assert.Equal(t, test.typ.String(), reflect.TypeOf(to).String())
		assert.Equal(t, test.length, reflect.ValueOf(to).Len())
	}
}

func TestIndirect(t *testing.T) {
	var a = 100
	var b *int

	tests := []struct {
		tag    string
		value  interface{}
		result interface{}
		isNil  bool
	}{
		{"t1", 100, 100, false},
		{"t2", &a, 100, false},
		{"t3", b, nil, true},
		{"t4", nil, nil, true},
	}

	for _, test := range tests {
		result, isNil := transformation.Indirect(test.value)
		assert.Equal(t, test.result, result, test.tag)
		assert.Equal(t, test.isNil, isNil, test.tag)
	}
}
