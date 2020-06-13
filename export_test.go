package transformation

var (
	ApplyTransformers   = applyTransformers
	CopyValue           = copyValue
	MakePtr             = makePtr
	MakeSliceFrom       = makeSliceFrom
	Assign              = assign
	ToConcreteSlice     = toConcreteSlice
	SliceHasSameType    = sliceHasSameType
	Indirect            = indirect
	MapHasSameValueType = mapHasSameValueType
	MapHasSameKeyType   = mapHasSameKeyType
	ToConcreteMap       = toConcreteMap
)
