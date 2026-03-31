package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bbsify-landed/clog"
	hw "github.com/bbsify-landed/heartwood/pkg/heartwood"

	s "github.com/bbsify-landed/heartwood/bin/schema-example/schema"
)

func main() {
	ctx := context.Background()
	ctx = clog.WithLogger(ctx, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	app := hw.New()

	s.RegisterHealthCheck(app, func(ctx context.Context, req *s.HealthCheckRequest) (*s.HealthCheckResponse, error) {
		if req.AreYouHealthy != "are you healthy?" {
			return nil, fmt.Errorf("expected 'are you healthy?' in field 'are_you_healthy'")
		}
		return &s.HealthCheckResponse{
			Healthy: "healthy",
		}, nil
	})

	if err := hw.ListenAndServe(app, ctx, ":9000"); err != nil {
		clog.Error(ctx, "error listening", "err", err)
	}
}
