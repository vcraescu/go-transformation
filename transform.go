package transformation

import (
	"fmt"
	"reflect"
)

type (
	FieldTransformer struct {
		from         interface{}
		to           interface{}
		transformers []Transformer
	}

	Transformer interface {
		Transform(from interface{}) (interface{}, error)
	}

	Transformable interface {
		Transform() (interface{}, error)
	}

	TransformFunc func(from interface{}) (interface{}, error)
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
	v := reflect.ValueOf(to)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%T must be a pointer", to))
	}

	if !v.Elem().CanSet() {
		panic(fmt.Errorf("%T must be a addressable", to))
	}

	tmpTo, err := transform(from, transformers...)
	if err != nil {
		return nil
	}

	if tmpTo == nil {
		return nil
	}

	mustCopyValue(tmpTo, to)

	return nil
}

func transform(from interface{}, transformers ...Transformer) (interface{}, error) {
	tmpTo := from
	if v, ok := from.(Transformable); ok {
		var err error
		if tmpTo, err = v.Transform(); err != nil {
			return nil, err
		}
	}

	tmpTo, err := applyTransformers(tmpTo, transformers...)
	if err != nil {
		return nil, err
	}

	return tmpTo, nil
}

func applyTransformers(from interface{}, transformers ...Transformer) (interface{}, error) {
	from, _ = indirect(from)

	if len(transformers) == 0 {
		return from, nil
	}

	var to interface{}
	var err error
	for _, transformer := range transformers {
		to, err = transformer.Transform(from)
		if err != nil {
			return nil, err
		}
		from = to
	}

	return to, nil
}

func Field(from interface{}, to interface{}, transformers ...Transformer) *FieldTransformer {
	return &FieldTransformer{
		from:         from,
		to:           to,
		transformers: transformers,
	}
}

func By(fn TransformFunc) *inlineTransformer {
	return &inlineTransformer{fn: fn}
}

func Each(transformers ...Transformer) *eachTransformer {
	return &eachTransformer{transformers: transformers}
}

func Default(value interface{}) *inlineTransformer {
	return By(func(from interface{}) (interface{}, error) {
		ifrom, isNil := indirect(from)
		if isNil || reflect.ValueOf(ifrom).IsZero() {
			return value, nil
		}

		return from, nil
	})
}
