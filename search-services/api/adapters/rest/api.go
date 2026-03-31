package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/api/core"
)

type pingResponse struct {
	Replies map[string]string `json:"replies"`
}

type updateResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type statusResponse struct {
	Status core.UpdateStatus `json:"status"`
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

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			log.Error("error while update", "error", err)
			if errors.Is(err, core.ErrAlreadyExists) {
				http.Error(w, err.Error(), http.StatusAccepted)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		stats, err := updater.Stats(ctx)
		if err != nil {
			log.Error("cannot get stats", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := updateResponse{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("cannot encode stats", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("cannot get status", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		log.Info("get status:", "status", status)

		reply := statusResponse{
			Status: status,
		}
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode status", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("drop failed", "error", err)
			http.Error(w, "drop failed", http.StatusInternalServerError)
			return
		}
	}
}
