package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
	"github.com/saubuny/bootdev-rss/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbUrl := os.Getenv("CONN")
	db, err := sql.Open("postgres", dbUrl)
	dbQueries := database.New(db)

	cfg := apiConfig{DB: dbQueries}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /v1/healthz", healthHandler)
	serveMux.HandleFunc("GET /v1/err", errorHandler)
	serveMux.HandleFunc("POST /v1/users", cfg.createUserHandler)
	serveMux.HandleFunc("GET /v1/users", cfg.getUserByApiKeyHandler)
	serveMux.HandleFunc("POST /v1/feeds", cfg.middlewareAuth(cfg.createFeedHandler))
	serveMux.HandleFunc("GET /v1/feeds", cfg.getAllFeedsHandler)
	serveMux.HandleFunc("POST /v1/feed_follows", cfg.middlewareAuth(cfg.createFeedFollowHandler))
	serveMux.HandleFunc("GET /v1/feed_follows", cfg.middlewareAuth(cfg.getFeedFollowsHandler))
	serveMux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", cfg.middlewareAuth((cfg.deleteFeedFollowHandler)))

	go cfg.feedFetchWorker()
	server := http.Server{Handler: serveMux, Addr: "localhost:" + port}
	fmt.Println("[Info] Starting server on port", 8080)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Info] Server ended")
}
