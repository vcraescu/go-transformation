package transformation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	Trim     = TrimTransformer{}
	Copy     = CopyTransformer{}
	Money100 = MoneyTransformer{division: 100}
	ToString = ToStringTransformer{}
	Reverse  = ReverseTransformer{}
)

type (
	FieldTransformer struct {
		from         interface{}
		to           interface{}
		transformers []Transformer
	}

	Transformer interface {
		Transform(from, to interface{}) error
	}

	Transformable interface {
		Transform(to interface{}) error
	}

	TransformFunc func(from interface{}) (interface{}, error)

	CopyTransformer struct{}

	TrimTransformer struct{}

	MoneyTransformer struct {
		division int
	}

	ReverseTransformer struct{}

	ToStringTransformer struct{}

	inlineTransformer struct {
		fn TransformFunc
	}

	eachTransformer struct {
		transformers []Transformer
	}
)

func TransformStruct(from interface{}, fields ...*FieldTransformer) error {
	value := reflect.ValueOf(from)
	if value.Kind() != reflect.Ptr || (!value.IsNil() && value.Elem().Kind() != reflect.Struct) {
		return fmt.Errorf("must be a pointer to a struct but got %T", from)
	}

	if value.IsNil() {
		return nil
	}
	value = value.Elem()

	errs := Errors{}

	for _, field := range fields {
		fv := reflect.ValueOf(field.from)
		if fv.Kind() != reflect.Ptr {
			return fmt.Errorf("from field expected to be a pointer but got %T", field.from)
		}
		ft := findStructField(value, fv)
		if ft == nil {
			return fmt.Errorf("from field not found")
		}

		if err := Transform(field.from, field.to, field.transformers...); err != nil {
			errs[ft.Name] = err
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func Transform(from interface{}, to interface{}, transformers ...Transformer) error {
	if err := applyTransformers(from, to, transformers...); err != nil {
		return err
	}

	fromValue := reflect.ValueOf(from)
	if (fromValue.Kind() == reflect.Ptr || fromValue.Kind() == reflect.Interface) && fromValue.IsNil() {
		return nil
	}

	if v, ok := from.(Transformable); ok {
		return v.Transform(to)
	}

	return nil
}

func applyTransformers(from interface{}, to interface{}, transformers ...Transformer) error {
	from, _ = Indirect(from)
	if len(transformers) == 0 {
		transformers = append(transformers, Copy)
	}

	l := len(transformers)
	for _, t := range transformers[:l-1] {
		v := reflect.New(reflect.TypeOf(to))
		v.Elem().Set(reflect.ValueOf(to))
		tmpTo := v.Elem().Interface()
		if err := t.Transform(from, tmpTo); err != nil {
			return err
		}
		from, _ = Indirect(tmpTo)
	}

	vto := reflect.ValueOf(to)
	if vto.Kind() == reflect.Ptr && vto.Elem().Kind() == reflect.Ptr && !vto.Elem().IsNil() {
		to = vto.Elem().Interface()
	}

	from = reflect.Indirect(reflect.ValueOf(from)).Interface()

	return transformers[l-1].Transform(from, to)
}

func Field(from interface{}, to interface{}, transformers ...Transformer) *FieldTransformer {
	return &FieldTransformer{
		from:         from,
		to:           to,
		transformers: transformers,
	}
}

func (t CopyTransformer) Transform(from interface{}, to interface{}) error {
	MustCopyValue(from, to)

	return nil
}

func (t TrimTransformer) Transform(from interface{}, to interface{}) error {
	ifrom, _ := Indirect(from)
	if ifrom == nil {
		return nil
	}

	s := ifrom.(string)
	s = strings.Trim(s, " \n\t")

	return Copy.Transform(&s, to)
}

func (t MoneyTransformer) Transform(from interface{}, to interface{}) error {
	ifrom, _ := Indirect(from)
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

	return Copy.Transform(&amount, to)
}

func (t ToStringTransformer) Transform(from interface{}, to interface{}) error {
	s := fmt.Sprintf("%v", from)

	return Copy.Transform(&s, to)
}

func (t inlineTransformer) Transform(from interface{}, to interface{}) error {
	ifrom, _ := Indirect(from)
	tmpTo, err := t.fn(ifrom)
	if err != nil {
		return err
	}

	MustCopyValue(tmpTo, to)

	return nil
}

func (t eachTransformer) Transform(from interface{}, to interface{}) error {
	errs := Errors{}

	fromValue := reflect.ValueOf(from)
	toType := reflect.TypeOf(to).Elem()
	switch fromValue.Kind() {
	case reflect.Slice, reflect.Array:
		toSlice := reflect.MakeSlice(toType, 0, 0)
		for i := 0; i < fromValue.Len(); i++ {
			val := fromValue.Index(i).Interface()
			el := reflect.New(toType.Elem()).Interface()
			if err := Transform(val, el, t.transformers...); err != nil {
				errs[strconv.Itoa(i)] = err
				continue
			}

			toSlice = reflect.Append(toSlice, reflect.ValueOf(el).Elem())
		}
		reflect.ValueOf(to).Elem().Set(toSlice)
	default:
		return errors.New("must be an iterable (slice or array)")
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
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

func (t ReverseTransformer) Transform(from, to interface{}) error {
	s := from.(string)
	var sb strings.Builder

	for i := len(s) - 1; i >= 0; i-- {
		sb.WriteRune(rune(s[i]))
	}

	return Copy.Transform(sb.String(), to)
}

func By(fn TransformFunc) *inlineTransformer {
	return &inlineTransformer{fn: fn}
}

func Each(transformers ...Transformer) *eachTransformer {
	return &eachTransformer{transformers: transformers}
}
