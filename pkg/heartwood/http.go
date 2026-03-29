package heartwood

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/bbsify-landed/clog"
)

type BuiltInError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewServeMux(app *App, ctx context.Context) *http.ServeMux {
	mu := http.NewServeMux()

	for path := range app.handlers {
		clog.Info(ctx, "registering", "path", path)
		mu.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			err := Handle(app, ctx, r.Method, path, r.Body, w)

			var hwErr *HeartwoodError
			if err == io.EOF {
				clog.Debug(ctx, "client dropped connection")
			} else if err != nil {
				if !errors.As(err, &hwErr) {
					clog.Error(ctx, "did not handle error", "err", err)
					hwErr = &HeartwoodError{
						StatusCode: 500,
						Err:        errors.New("internal server error"),
					}
				}

				w.WriteHeader(hwErr.StatusCode)
				if err := hwErr.Serialize(w); err != nil {
					clog.Error(ctx, "failed to serialize outgoing error", "err", err)
				}
			}
		})
	}

	return mu
}

func ListenAndServe(app *App, ctx context.Context, address string) error {
	clog.Info(ctx, "listening", "address", address)

	mu := NewServeMux(app, ctx)

	return http.ListenAndServe(address, mu)
}
