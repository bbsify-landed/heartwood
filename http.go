package heartwood

import (
	"context"
	"errors"
	"net/http"

	"github.com/bbsify-landed/clog"
)

// BuiltInError is a simple JSON-serializable error body used internally.
type BuiltInError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NewServeMux builds an [http.ServeMux] from all handlers registered on app.
// Errors returned by handlers are translated to appropriate HTTP status codes:
// [*HeartwoodError] uses its StatusCode, other errors become 500.
func NewServeMux(app *App, ctx context.Context) *http.ServeMux {
	mu := http.NewServeMux()

	for path := range app.handlers {
		clog.Info(ctx, "registering", "path", path)
		mu.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			err := Handle(app, ctx, r.Method, path, r.Body, w)

			var hwErr *HeartwoodError
			if err != nil {
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

// ListenAndServe is a convenience wrapper that builds a serve mux from app
// and starts an HTTP server on the given address.
func ListenAndServe(app *App, ctx context.Context, address string) error {
	clog.Info(ctx, "listening", "address", address)

	mu := NewServeMux(app, ctx)

	return http.ListenAndServe(address, mu)
}
