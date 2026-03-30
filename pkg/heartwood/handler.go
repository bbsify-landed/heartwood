package heartwood

import (
	"context"
	"io"
)

type Serializable interface {
	Serialize(w io.Writer) error
}

type Deserializable[T any] interface {
	*T
	Deserialize(r io.Reader) error
	Validate() error
}

type Handler[R any, D Deserializable[R], RR Serializable] func(ctx context.Context, req D) (error, RR)
