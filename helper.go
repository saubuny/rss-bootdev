package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/saubuny/bootdev-rss/internal/database"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code >= 500 {
		fmt.Printf("Responding with 5XX error: %s", msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

type Feed struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastFetchedAt time.Time
}

func databaseFeedToFeed(feed database.Feed) Feed {
	return Feed{
		ID:            feed.ID,
		UserID:        feed.UserID,
		CreatedAt:     feed.CreatedAt,
		UpdatedAt:     feed.UpdatedAt,
		LastFetchedAt: feed.LastFetchedAt.Time,
	}
}
