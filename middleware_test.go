package heartwood_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bbsify-landed/clog"
	hw "github.com/bbsify-landed/heartwood"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareOrder(t *testing.T) {
	var order []string

	app := hw.New()
	hw.Use(app, "POST", "/health", func(ctx context.Context, req *Foo) (*Baz, error) {
		return &Baz{Ble: "bob"}, nil
	})

	app.With(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "first-before")
				next.ServeHTTP(w, r)
				order = append(order, "first-after")
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "second-before")
				next.ServeHTTP(w, r)
				order = append(order, "second-after")
			})
		},
	)

	mux := hw.NewServeMux(app, t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/health", bytes.NewBufferString(`{"bar":"alice"}`))
	mux.ServeHTTP(w, r)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, []string{"first-before", "second-before", "second-after", "first-after"}, order)
}

func TestRequestLogger(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	ctx := clog.WithLogger(t.Context(), logger)

	app := hw.New()
	hw.Use(app, "POST", "/health", func(ctx context.Context, req *Foo) (*Baz, error) {
		return &Baz{Ble: "bob"}, nil
	})
	app.With(hw.RequestLogger(ctx))

	mux := hw.NewServeMux(app, ctx)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/health", bytes.NewBufferString(`{"bar":"alice"}`))
	mux.ServeHTTP(w, r)

	assert.Equal(t, 200, w.Code)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "request")
	assert.Contains(t, logOutput, "method=POST")
	assert.Contains(t, logOutput, "path=/health")
	assert.Contains(t, logOutput, "status=200")
	assert.Contains(t, logOutput, "duration=")
}

func TestRequestLoggerCapturesErrorStatus(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	ctx := clog.WithLogger(t.Context(), logger)

	app := hw.New()
	hw.Use(app, "POST", "/health", func(ctx context.Context, req *Foo) (*Baz, error) {
		return nil, hw.Error(422, assert.AnError)
	})
	app.With(hw.RequestLogger(ctx))

	mux := hw.NewServeMux(app, ctx)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/health", bytes.NewBufferString(`{"bar":"alice"}`))
	mux.ServeHTTP(w, r)

	assert.Equal(t, 422, w.Code)
	assert.Contains(t, logBuf.String(), "status=422")
}
