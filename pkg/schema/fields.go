package schema

// String creates a string field.
func String(name string) Field {
	return Field{Name: name, Type: StringType}
}

// Int creates an int field.
func Int(name string) Field {
	return Field{Name: name, Type: IntType}
}

// Int64 creates an int64 field.
func Int64(name string) Field {
	return Field{Name: name, Type: Int64Type}
}

// Float64 creates a float64 field.
func Float64(name string) Field {
	return Field{Name: name, Type: Float64Type}
}

// Bool creates a bool field.
func Bool(name string) Field {
	return Field{Name: name, Type: BoolType}
}

// Ref creates a field that references another Definition (nested message).
func Ref(name string, def *Definition) Field {
	return Field{Name: name, Type: RefType, RefName: def.Name}
}

// Slice creates a field that is a slice of scalar values.
func Slice(name string, elemType FieldType) Field {
	return Field{Name: name, Type: SliceType, ElemType: elemType}
}

// SliceOf creates a field that is a slice of another Definition (slice of messages).
func SliceOf(name string, def *Definition) Field {
	return Field{Name: name, Type: SliceRefType, RefName: def.Name}
}

// Required marks the field as required for validation.
func (f Field) Required() Field {
	f.IsRequired = true
	return f
}

// MinLength sets the minimum length constraint (for strings and slices).
func (f Field) MinLength(n int) Field {
	f.MinLen = &n
	return f
}

// MaxLength sets the maximum length constraint (for strings and slices).
func (f Field) MaxLength(n int) Field {
	f.MaxLen = &n
	return f
}

// MinValue sets the minimum value constraint (for numeric types).
func (f Field) MinValue(n float64) Field {
	f.Min = &n
	return f
}

// MaxValue sets the maximum value constraint (for numeric types).
func (f Field) MaxValue(n float64) Field {
	f.Max = &n
	return f
}
