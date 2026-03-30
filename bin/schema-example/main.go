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

	s.RegisterHealthCheck(app, func(ctx context.Context, req *s.HealthCheckRequest) (error, *s.HealthCheckResponse) {
		if req.AreYouHealthy != "are you healthy?" {
			return fmt.Errorf("expected 'are you healthy?' in field 'are_you_healthy'"), nil
		}
		return nil, &s.HealthCheckResponse{
			Healthy: "healthy",
		}
	})

	if err := hw.ListenAndServe(app, ctx, ":9000"); err != nil {
		clog.Error(ctx, "error listening", "err", err)
	}
}
