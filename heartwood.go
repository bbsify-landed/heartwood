// Package heartwood is a lightweight HTTP framework for building type-safe
// request/response handlers in Go using generics.
//
// Heartwood separates transport concerns from business logic. You define
// request and response types that implement [Serializable] and [Deserializable],
// then write handler functions with concrete types. The framework handles
// deserialization, validation, and serialization automatically.
//
// A companion code generator (cmd/hwgen) can produce handler boilerplate and
// typed HTTP clients from declarative schema definitions.
package heartwood

import (
	"context"
	"errors"
	"io"
)

type handler func(ctx context.Context, r io.Reader, w io.Writer) error

// App is the central registry for heartwood handlers. Create one with [New]
// and register handlers with [Use].
type App struct {
	handlers map[string]map[string]handler
}

// New creates a new [App] with an empty handler registry.
func New() *App {
	return &App{
		handlers: map[string]map[string]handler{},
	}
}

// Use registers a typed handler for the given HTTP method and path on app.
// The type parameters wire up automatic deserialization, validation, and
// serialization: the request body is decoded into D, validated, passed to
// the handler, and the returned RR is serialized to the response writer.
func Use[R any, D Deserializable[R], RR Serializable](
	app *App,
	method string,
	path string,
	h Handler[R, D, RR],
) {
	if _, ok := app.handlers[path]; !ok {
		app.handlers[path] = map[string]handler{}
	}

	app.handlers[path][method] = func(ctx context.Context, r io.Reader, w io.Writer) error {
		req := D(new(R))
		if err := req.Deserialize(r); err != nil {
			return err
		}

		if err := req.Validate(); err != nil {
			return Error(400, err)
		}

		res, err := h(ctx, req)
		if err != nil {
			return err
		}

		return res.Serialize(w)
	}
}

// Handle dispatches a request to the handler registered for the given method
// and path. It returns a [*HeartwoodError] with status 405 if no handler
// matches the method.
func Handle(app *App, ctx context.Context, method string, path string, r io.Reader, w io.Writer) error {
	methods := app.handlers[path]
	handler, ok := methods[method]
	if !ok {
		return Error(405, errors.New("method not allowed"))
	}
	if err := handler(ctx, r, w); err != nil {
		return err
	}

	return nil
}
