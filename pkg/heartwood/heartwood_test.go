package heartwood_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	hw "github.com/bbsify-landed/heartwood/pkg/heartwood"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	app := SimpleApp()

	r, err := marshalReader(&Foo{Bar: "alice"})
	require.Nil(t, err, err)

	w := bytes.NewBuffer(nil)

	ctx := t.Context()
	err = hw.Handle(app, ctx, "POST", "/health", r, w)
	require.Nil(t, err, err)

	d := json.NewDecoder(w)
	var b Baz
	err = d.Decode(&b)
	require.Nil(t, err, err)
	assert.Equal(t, b.Ble, "bob")
}

func TestSimpleError(t *testing.T) {
	app := SimpleApp()

	r, err := marshalReader(&Foo{Bar: "bob"})
	require.Nil(t, err, err)

	w := bytes.NewBuffer(nil)

	ctx := t.Context()
	err = hw.Handle(app, ctx, "POST", "/health", r, w)

	var hwError *hw.HeartwoodError
	require.True(t, errors.As(err, &hwError))
	assert.Equal(t, hwError.StatusCode, 400)
}

func TestHeartwoodError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	hwErr := hw.Error(500, inner)
	assert.Equal(t, inner, errors.Unwrap(hwErr))
}

func TestClientError(t *testing.T) {
	ce := &hw.ClientError{
		StatusCode: 403,
		Err:        "forbidden",
	}
	assert.Equal(t, "[403] forbidden", ce.Error())
	assert.NoError(t, ce.Validate())

	r := bytes.NewReader([]byte(`{"status_code": 403, "error": "forbidden"}`))
	ce2 := &hw.ClientError{}
	err := ce2.Deserialize(r)
	assert.NoError(t, err)
	assert.Equal(t, 403, ce2.StatusCode)
	assert.Equal(t, "forbidden", ce2.Err)

	// Test invalid JSON
	r2 := bytes.NewReader([]byte(`{invalid}`))
	err2 := ce2.Deserialize(r2)
	assert.Error(t, err2)
}

func TestNewServeMux_Errors(t *testing.T) {
	app := SimpleApp()
	// Add an endpoint that returns a generic error
	hw.Use(app, "POST", "/error", func(ctx context.Context, req *Foo) (*Baz, error) {
		return nil, errors.New("generic error")
	})
	// Add an endpoint that returns EOF
	hw.Use(app, "POST", "/eof", func(ctx context.Context, req *Foo) (*Baz, error) {
		return nil, io.EOF
	})

	mux := hw.NewServeMux(app, t.Context())

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "Path not found",
			method:     "POST",
			path:       "/not-found",
			body:       `{"bar":"alice"}`,
			wantStatus: 404, // mux.ServeHTTP returns 404 for unknown paths
		},
		{
			name:       "Method not allowed",
			method:     "GET",
			path:       "/health",
			body:       `{"bar":"alice"}`,
			wantStatus: 405,
		},
		{
			name:       "Generic error (500)",
			method:     "POST",
			path:       "/error",
			body:       `{"bar":"alice"}`,
			wantStatus: 500,
		},
		{
			name:       "IO EOF",
			method:     "POST",
			path:       "/eof",
			body:       `{"bar":"alice"}`,
			wantStatus: 200, // ServeMux treats io.EOF as success (client dropped)
		},
		{
			name:       "Invalid Request Body",
			method:     "POST",
			path:       "/health",
			body:       `{invalid}`,
			wantStatus: 500, // Handle returns a generic error on deserialize failure, which NewServeMux turns into 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			mux.ServeHTTP(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

type failWriter struct{}

func (fw *failWriter) Header() http.Header       { return http.Header{} }
func (fw *failWriter) Write([]byte) (int, error) { return 0, errors.New("write failure") }
func (fw *failWriter) WriteHeader(int)           {}

func TestNewServeMux_SerializationError(t *testing.T) {
	app := SimpleApp()
	mux := hw.NewServeMux(app, t.Context())

	// Trigger an error that will be serialized, then fail the serialization
	r := httptest.NewRequest("POST", "/health", bytes.NewBufferString(`{"bar":"bob"}`))
	w := &failWriter{}

	// This won't panic but we expect it to log (which we aren't easily checking here)
	// We just want to cover the branch where Serialize(w) fails
	mux.ServeHTTP(w, r)
}

func TestListenAndServe_Error(t *testing.T) {
	app := SimpleApp()
	// An invalid address to cause ListenAndServe to fail.
	err := hw.ListenAndServe(app, t.Context(), "invalid:address")
	assert.Error(t, err)
}
