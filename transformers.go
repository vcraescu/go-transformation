package transformation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	Trim      = TrimTransformer{}
	Money100  = MoneyTransformer{division: 100}
	ToString  = ToStringTransformer{}
	Reverse   = ReverseTransformer{}
	UpperCase = UpperCaseTransformer{}
	DownCase  = DownCaseTransformer{}
)

type (
	TrimTransformer struct{}

	MoneyTransformer struct {
		division int
	}

	ReverseTransformer struct{}

	UpperCaseTransformer struct{}

	DownCaseTransformer struct{}

	ToStringTransformer struct{}

	inlineTransformer struct {
		fn TransformFunc
	}

	eachTransformer struct {
		transformers []Transformer
	}
)

func (t TrimTransformer) Transform(from interface{}) (interface{}, error) {
	v, err := ToString.Transform(from)
	if err != nil {
		return v, err
	}

	s := v.(string)
	return strings.Trim(s, " \n\t"), nil
}

func (t UpperCaseTransformer) Transform(from interface{}) (interface{}, error) {
	v, err := ToString.Transform(from)
	if err != nil {
		return v, err
	}

	s := v.(string)
	return strings.ToUpper(s), nil
}

func (t DownCaseTransformer) Transform(from interface{}) (interface{}, error) {
	v, err := ToString.Transform(from)
	if err != nil {
		return v, err
	}

	s := v.(string)
	return strings.ToLower(s), nil
}

func (t MoneyTransformer) Transform(from interface{}) (interface{}, error) {
	ifrom, _ := indirect(from)
	var amount int64

	switch v := ifrom.(type) {
	case float32:
		amount = int64(v * float32(t.division))
		break
	case float64:
		amount = int64(v * float64(t.division))
		break
	default:
		panic(fmt.Errorf("invalid value type; expected float32 or float64 type but got %T", ifrom))
	}

	return amount, nil
}

func (t ToStringTransformer) Transform(from interface{}) (interface{}, error) {
	ifrom, isNil := indirect(from)
	if isNil {
		return "", nil
	}

	s := fmt.Sprintf("%v", ifrom)

	return s, nil
}

func (t inlineTransformer) Transform(from interface{}) (interface{}, error) {
	ifrom, _ := indirect(from)

	return t.fn(ifrom)
}

func (t eachTransformer) Transform(from interface{}) (interface{}, error) {
	errs := Errors{}

	fromValue := reflect.ValueOf(from)
	var to interface{}
	var err error
	switch fromValue.Kind() {
	case reflect.Slice, reflect.Array:
		sl := make([]interface{}, fromValue.Len())
		for i := 0; i < fromValue.Len(); i++ {
			el := fromValue.Index(i).Interface()
			sl[i], err = transform(el, t.transformers...)
			if err != nil {
				errs[strconv.Itoa(i)] = err
				continue
			}
		}

		to = toConcreteSlice(sl)
	case reflect.Map:
		m := make(map[interface{}]interface{})
		iter := fromValue.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			m[k.Interface()], err = transform(v.Interface(), t.transformers...)
			if err != nil {
				errs[k.String()] = err
				continue
			}
		}

		to = toConcreteMap(m)
	default:
		return nil, errors.New("must be an iterable (slice, array, map)")
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return to, nil
}

func (t eachTransformer) getInterface(value reflect.Value) interface{} {
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return value.Elem().Interface()
	default:
		return value.Interface()
	}
}

func (t ReverseTransformer) Transform(from interface{}) (interface{}, error) {
	ifrom, isNil := indirect(from)
	if isNil {
		return nil, nil
	}

	s := ifrom.(string)
	var sb strings.Builder

	for i := len(s) - 1; i >= 0; i-- {
		sb.WriteRune(rune(s[i]))
	}

	return sb.String(), nil
}
