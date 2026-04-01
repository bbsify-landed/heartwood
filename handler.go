package heartwood

import (
	"context"
	"io"
)

// Serializable is implemented by response types that can encode themselves
// to a writer (typically as JSON).
type Serializable interface {
	Serialize(w io.Writer) error
}

// Deserializable is implemented by request types that can decode themselves
// from a reader and validate the result. The pointer-to-T constraint lets
// [Use] instantiate the concrete request type automatically.
type Deserializable[T any] interface {
	*T
	Deserialize(r io.Reader) error
	Validate() error
}

// Handler is the signature for a typed heartwood endpoint. It receives a
// deserialized and validated request and returns a serializable response.
type Handler[R any, D Deserializable[R], RR Serializable] func(ctx context.Context, req D) (RR, error)
