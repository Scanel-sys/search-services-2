package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
)

type pingResponse struct {
	Replies map[string]string `json:"replies"`
}

type normResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		replies := make(map[string]string, len(pingers))

		for name, pinger := range pingers {
			if err := pinger.Ping(ctx); err != nil {
				log.Error("service unavailable", "service", name, "error", err)
				replies[name] = "unavailable"
				continue
			}

			replies[name] = "ok"
		}

		resp := pingResponse{
			Replies: replies,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("cannot encode response", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
}

func NewNormHandler(log *slog.Logger, normalizer core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "phrase is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		out, err := normalizer.Norm(ctx, phrase)
		if err != nil {

			if status.Code(err) == codes.ResourceExhausted {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			log.Error("normalization failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resp := normResponse{
			Words: out,
			Total: len(out),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("cannot encode response", "error", err)
			return
		}
	}
}
