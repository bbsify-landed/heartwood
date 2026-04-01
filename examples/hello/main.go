// Command hello demonstrates a minimal heartwood HTTP server with a single
// typed endpoint.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/bbsify-landed/clog"
	hw "github.com/bbsify-landed/heartwood"
)

// GreetRequest is the request body for POST /greet.
type GreetRequest struct {
	Name string `json:"name"`
}

func (r *GreetRequest) Deserialize(rd io.Reader) error { return json.NewDecoder(rd).Decode(r) }
func (r *GreetRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// GreetResponse is the response body for POST /greet.
type GreetResponse struct {
	Message string `json:"message"`
}

func (r *GreetResponse) Serialize(w io.Writer) error { return json.NewEncoder(w).Encode(r) }

func main() {
	ctx := context.Background()
	ctx = clog.WithLogger(ctx, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	app := hw.New()

	hw.Use(app, "POST", "/greet", func(ctx context.Context, req *GreetRequest) (*GreetResponse, error) {
		return &GreetResponse{
			Message: "Hello, " + req.Name + "!",
		}, nil
	})

	if err := hw.ListenAndServe(app, ctx, ":9000"); err != nil {
		clog.Error(ctx, "server error", "err", err)
		os.Exit(1)
	}
}
