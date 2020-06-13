package transformation_test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/vcraescu/go-transformation"
	"testing"
	"time"
)

type (
	Address struct {
		Line1 string
		Line2 string
		No    *int
	}

	Person struct {
		ID        *int
		FirstName string
		LastName  *string
		Birthdate *time.Time
		Age       *int
		Addresses []string
	}
)

func TestApplyTransformers1(t *testing.T) {
	from := "this is a test  "
	var to interface{}
	to, err := transformation.ApplyTransformers(from, transformation.Trim)
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", to.(string))
	}

	from = "foobar   "
	to, err = transformation.ApplyTransformers(from, transformation.Trim, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "raboof", to)
	}
}

func TestApplyTransformers2(t *testing.T) {
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to, err := transformation.ApplyTransformers(from)
	if assert.NoError(t, err) {
		tm, ok := to.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, from.Unix(), tm.Unix())
	}
}

func TestApplyTransformers3(t *testing.T) {
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to, err := transformation.ApplyTransformers(&from)
	if assert.NoError(t, err) {
		tm, ok := to.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, from.Unix(), tm.Unix())
	}
}

func TestApplyTransformers4(t *testing.T) {
	var to interface{}
	from := 1200
	to, err := transformation.ApplyTransformers(from, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "0021", to.(string))
	}
}

func TestApplyTransformers5(t *testing.T) {
	var to interface{}
	var from *string
	to, err := transformation.ApplyTransformers(from, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Zero(t, to)
	}
}

func TestApplyTransformers6(t *testing.T) {
	from := []string{"1  ", "2   ", "  3"}
	var to interface{}
	to, err := transformation.ApplyTransformers(from, transformation.Each(transformation.Trim))
	if assert.NoError(t, err) {
		sl, ok := to.([]string)
		assert.True(t, ok)
		assert.Len(t, sl, len(from))
	}
}

func TestApplyTransformers7(t *testing.T) {
	v1 := "1   "
	v2 := "2   "
	v3 := "3   "
	from := []*string{nil, &v1, &v2, &v3}
	var to interface{}
	to, err := transformation.ApplyTransformers(from, transformation.Each(transformation.Trim))
	if assert.NoError(t, err) {
		sl, ok := to.([]string)
		assert.True(t, ok)
		assert.Len(t, sl, len(from))
	}
}

func TestApplyTransformers8(t *testing.T) {
	from := map[int]string{0: "1  ", 1: "2  ", 3: "3    "}
	var to interface{}
	to, err := transformation.ApplyTransformers(from, transformation.Each(transformation.Trim))
	spew.Dump(to)
	if assert.NoError(t, err) {
		m, ok := to.(map[int]string)
		assert.True(t, ok)
		assert.Len(t, m, len(from))
	}
}

func TestTransform(t *testing.T) {
	var to string
	from := 1200
	err := transformation.Transform(from, &to, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "0021", to)
	}
}

func TestTransformNil(t *testing.T) {
	var from *int
	var to string
	err := transformation.Transform(from, &to, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Zero(t, to)
	}
}

func TestTransformNil2(t *testing.T) {
	var from *int
	var to *string
	err := transformation.Transform(from, &to)
	if assert.NoError(t, err) {
		assert.Nil(t, to)
	}
}

func TestTransformSlice(t *testing.T) {
	from := []string{"1  ", "2   ", "  3"}
	var to []*string
	err := transformation.Transform(from, &to, transformation.Each(transformation.Trim))
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		assert.Equal(t, "1", *to[0])
		assert.Equal(t, "2", *to[1])
		assert.Equal(t, "3", *to[2])
	}
}

func TestTransformSlice2(t *testing.T) {
	v1 := "1   "
	v2 := "2   "
	v3 := "3   "
	from := []*string{nil, &v1, &v2, &v3}
	var to []*string
	err := transformation.Transform(&from, &to, transformation.Each(transformation.Trim))
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		assert.Equal(t, "", *to[0])
		assert.Equal(t, "1", *to[1])
		assert.Equal(t, "2", *to[2])
		assert.Equal(t, "3", *to[3])
	}
}

func TestTransformSlice3(t *testing.T) {
	v1 := time.Now()
	v2 := time.Now()
	v3 := time.Now()
	from := []*time.Time{&v1, &v2, &v3}
	type MyTime struct {
		Time string
	}

	var to []*MyTime
	err := transformation.Transform(
		&from,
		&to,
		transformation.Each(
			transformation.ToString,
			transformation.Trim,
			transformation.By(func(from interface{}) (interface{}, error) {
				return MyTime{Time: from.(string)}, nil
			}),
		),
	)
	if assert.NoError(t, err) {
		assert.Len(t, to, len(from))
		assert.Equal(t, v1.String(), to[0].Time)
		assert.Equal(t, v2.String(), to[1].Time)
		assert.Equal(t, v3.String(), to[2].Time)
	}
}

func TestTransformObject(t *testing.T) {
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var to interface{}
	err := transformation.Transform(from, &to)
	if assert.NoError(t, err) {
		tm, ok := to.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, from.Unix(), tm.Unix())
	}

	to = nil
	err = transformation.Transform(from, &to, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		s, ok := to.(string)
		assert.True(t, ok)
		expected, _ := transformation.Reverse.Transform(from.String())
		assert.Equal(t, expected, s)
	}
}

func TestTransformPtrObject(t *testing.T) {
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var to interface{}
	err := transformation.Transform(&from, &to)
	if assert.NoError(t, err) {
		tm, ok := to.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, from.Unix(), tm.Unix())
	}
}

func TestTransformPtr(t *testing.T) {
	var to *string
	from := 1200
	err := transformation.Transform(from, &to, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "0021", *to)
	}
}

func TestTransformable(t *testing.T) {
	lastName := "Doe"
	p := Person{
		FirstName: "John",
		LastName:  &lastName,
	}

	var to *string
	err := transformation.Transform(p, &to, transformation.Trim)
	if assert.NoError(t, err) {
		assert.Equal(t, "John Doe", *to)
	}
}

func TestTransformStruct(t *testing.T) {
	lastName := "   Doe   "
	birthdate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	addresses := []string{"Stree1   ", " Street2  "}
	from := Person{
		FirstName: "John    ",
		LastName:  &lastName,
		Birthdate: &birthdate,
		Addresses: addresses,
	}

	type Foo struct {
		ID        string
		FirstName *string
		LastName  string
		Birthdate *time.Time
		Age       string
		Addresses []*Address
	}

	to := Foo{}

	err := transformation.TransformStruct(
		&from,
		transformation.Field(&from.FirstName, &to.FirstName, transformation.Trim),
		transformation.Field(&from.LastName, &to.LastName, transformation.Trim),
		transformation.Field(&from.Birthdate, &to.Birthdate),
		transformation.Field(&from.Age, &to.Age, transformation.ToString),
		transformation.Field(
			&from.ID,
			&to.ID,
			transformation.Default(1000),
			transformation.ToString,
		),
		transformation.Field(
			&from.Addresses,
			&to.Addresses,
			transformation.Each(
				transformation.Trim,
				transformation.Reverse,
				transformation.By(func(from interface{}) (interface{}, error) {
					line1 := from.(string)

					return &Address{
						Line1: line1,
					}, nil
				}),
			),
		),
	)

	if assert.NoError(t, err) {
		if assert.NotNil(t, to.FirstName) {
			assert.Equal(t, "John", *to.FirstName)
		}
		assert.Equal(t, "Doe", to.LastName)
		if assert.NotNil(t, to.Birthdate) {
			assert.Equal(t, from.Birthdate.Unix(), to.Birthdate.Unix())
		}
		assert.Zero(t, to.Age)
		assert.Equal(t, "1000", to.ID)
		assert.Len(t, to.Addresses, len(addresses))
		assert.Equal(t, "1eertS", to.Addresses[0].Line1)
	}
}

func (p Person) Transform() (interface{}, error) {
	return fmt.Sprintf("%s %s   ", p.FirstName, *p.LastName), nil
}
