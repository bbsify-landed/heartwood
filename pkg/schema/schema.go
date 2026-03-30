package schema

// FieldType represents the type of a schema field.
type FieldType int

const (
	StringType  FieldType = iota
	IntType               // int
	Int64Type             // int64
	Float64Type           // float64
	BoolType              // bool
	RefType               // nested message reference
	SliceType             // slice of scalars
	SliceRefType          // slice of message references
)

// Field represents a single field in a request or response message.
type Field struct {
	Name       string    `json:"name"`
	Type       FieldType `json:"type"`
	RefName    string    `json:"ref_name,omitempty"`
	ElemType   FieldType `json:"elem_type,omitempty"`
	IsRequired bool      `json:"required,omitempty"`
	MinLen     *int      `json:"min_len,omitempty"`
	MaxLen     *int      `json:"max_len,omitempty"`
	Min        *float64  `json:"min,omitempty"`
	Max        *float64  `json:"max,omitempty"`
}

// Definition describes a single endpoint with its request and response types.
type Definition struct {
	Name      string  `json:"name"`
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	ReqFields []Field `json:"request"`
	ResFields []Field `json:"response"`
}
