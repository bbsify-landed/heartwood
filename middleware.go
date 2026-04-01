package heartwood

import (
	"context"
	"net/http"
	"time"

	"github.com/bbsify-landed/clog"
)

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wrote {
		sw.status = code
		sw.wrote = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wrote {
		sw.status = 200
		sw.wrote = true
	}
	return sw.ResponseWriter.Write(b)
}

// RequestLogger returns a [Middleware] that logs every request with method,
// path, status code, and duration using clog. The provided ctx must carry a
// clog logger (see [clog.WithLogger]).
func RequestLogger(ctx context.Context) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}

			next.ServeHTTP(sw, r)

			clog.Info(ctx, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration", time.Since(start),
			)
		})
	}
}
