package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
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
	Name          string
	UserID        uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastFetchedAt time.Time
}

func databaseFeedToFeed(feed database.Feed) Feed {
	return Feed{
		ID:            feed.ID,
		Name:          feed.Name,
		UserID:        feed.UserID,
		CreatedAt:     feed.CreatedAt,
		UpdatedAt:     feed.UpdatedAt,
		LastFetchedAt: feed.LastFetchedAt.Time,
	}
}

// Rss was generated 2024-09-04 09:22:05 by https://xml-to-go.github.io/ in Ukraine.
type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"atom,attr"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		Language      string `xml:"language"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Guid        string `xml:"guid"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

func fetchFromFeed(feedUrl string) (Rss, error) {
	resp, err := http.Get(feedUrl)
	if err != nil {
		return Rss{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Rss{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Rss{}, fmt.Errorf("Read body: %v", err)
	}

	var res Rss
	err = xml.Unmarshal(data, &res)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (cfg *apiConfig) feedFetchWorker() {
	for {
		log.Println("Fetching feeds from DB...")
		feeds, err := cfg.DB.GetNextFeedsToFetch(context.Background(), 10)

		if err != nil {
			log.Println("Error in feed fetch worker: " + err.Error())
			continue
		}

		log.Println("Processing feeds...")

		var wg sync.WaitGroup
		for _, feed := range feeds {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				rss, err := fetchFromFeed(url)
				if err != nil {
					log.Println("Error in feed fetch worker: " + err.Error())
				}
				fmt.Println("Feed processed: " + rss.Channel.Title)
			}(feed.Url)
		}

		wg.Wait()
		time.Sleep(60 * time.Second)
	}
}
