package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bbsify-landed/clog"
)

type HealthCheckRequest struct {
	AreYouHealthy string `json:"are_you_healthy"`
}

type HealthCheckResponse struct {
	Healthy string `json:"healthy"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func main() {
	ctx := context.Background()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clog.Info(ctx, "/health", "host", r.Host)

		d := json.NewDecoder(r.Body)
		var req HealthCheckRequest
		if err := d.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			e := json.NewEncoder(w)
			err := e.Encode(&ErrorResponse{
				Error:   "ERR_INVALID_REQUEST",
				Message: fmt.Sprintf("invalid request, expected json: %s", err.Error()),
			})
			if err != nil {
				clog.Error(ctx, "could not encode request", "err", err)
			}
			return
		}

		if req.AreYouHealthy != "are you healthy?" {
			w.WriteHeader(http.StatusBadRequest)
			e := json.NewEncoder(w)
			err := e.Encode(&ErrorResponse{
				Error:   "ERR_INVALID_REQUEST",
				Message: "expected 'are you healthy?' in 'are_you_healthy' field",
			})
			if err != nil {
				clog.Error(ctx, "could not encode request", "err", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		e := json.NewEncoder(w)
		err := e.Encode(&HealthCheckResponse{
			Healthy: "healthy",
		})
		if err != nil {
			clog.Error(ctx, "msg", "could not encode request", "err", err)
		}
	})

	if err := http.ListenAndServe(":9000", nil); err != nil {
		clog.Error(ctx, "error listening and serving", "error", err)
	}

	clog.Info(ctx, "server stopped")
}
