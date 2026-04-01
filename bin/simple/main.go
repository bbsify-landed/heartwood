package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/bbsify-landed/clog"
	hw "github.com/bbsify-landed/heartwood"
)

type HealthCheckRequest struct {
	AreYouHealthy string `json:"are_you_healthy"`
}

func (h *HealthCheckRequest) Deserialize(r io.Reader) error {
	h = &HealthCheckRequest{}
	d := json.NewDecoder(r)
	if err := d.Decode(h); err != nil {
		return err
	}

	return nil
}

func (h *HealthCheckRequest) Validate() error { return nil }

type HealthCheckResponse struct {
	Healthy string `json:"healthy"`
}

func (h *HealthCheckResponse) Serialize(w io.Writer) error {
	e := json.NewEncoder(w)
	if err := e.Encode(h); err != nil {
		return err
	}

	return nil
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func main() {
	ctx := context.Background()

	ctx = clog.WithLogger(ctx, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	app := hw.New()

	hw.Use(
		app,
		"POST",
		"/health",
		func(
			ctx context.Context,
			req *HealthCheckRequest,
		) (*HealthCheckResponse, error) {
			if req.AreYouHealthy != "are you healthy?" {
				return nil, fmt.Errorf("expected 'are you healthy?' in field 'are_you_healthy'")
			}

			return &HealthCheckResponse{
				Healthy: "healthy",
			}, nil
		},
	)

	if err := hw.ListenAndServe(app, ctx, ":9000"); err != nil {
		clog.Error(ctx, "error")
	}

	clog.Info(ctx, "server stopped")
}
