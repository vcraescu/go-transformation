package transformation_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/vcraescu/transformation"
	"testing"
)

func TestApplyTransformers(t *testing.T) {
	from := "this is a test  "
	var to string
	err := transformation.ApplyTransformers(from, &to, transformation.Trim)
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", to)
	}

	from = "foobar   "
	err = transformation.ApplyTransformers(from, &to, transformation.Trim, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "raboof", to)
	}
}

func TestApplyTransformersPtr(t *testing.T) {
	var to *string
	from := "foobar   "
	err := transformation.ApplyTransformers(from, &to, transformation.Trim)
	if assert.NoError(t, err) {
		assert.Equal(t, "foobar", *to)
	}
}

func TestApplyTransformersPtrMultiple(t *testing.T) {
	var to *string
	from := "foobar   "
	err := transformation.ApplyTransformers(from, &to, transformation.Trim, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "raboof", *to)
	}
}

func TestApplyTransformersDifferentTypes(t *testing.T) {
	var to string
	from := 1200
	err := transformation.ApplyTransformers(from, &to, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "0021", to)
	}
}

func TestApplyTransformersDifferentTypesPtr(t *testing.T) {
	var to *string
	from := 1200
	err := transformation.ApplyTransformers(from, &to, transformation.ToString, transformation.Reverse)
	if assert.NoError(t, err) {
		assert.Equal(t, "0021", *to)
	}
}
