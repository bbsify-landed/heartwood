package heartwood

import (
	"context"
	"errors"
	"io"
)

type handler func(ctx context.Context, r io.Reader, w io.Writer) error

type App struct {
	handlers map[string]map[string]handler
}

func New() *App {
	return &App{
		handlers: map[string]map[string]handler{},
	}
}

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

		if v, ok := any(req).(Validatable); ok {
			if err := v.Validate(); err != nil {
				return Error(400, err)
			}
		}

		if err, res := h(ctx, req); err != nil {
			return err
		} else {
			if err := res.Serialize(w); err != nil {
				return err
			}

			return nil
		}
	}
}

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
