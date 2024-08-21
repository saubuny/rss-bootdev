package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/saubuny/bootdev-rss/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type res struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, res{Status: "OK"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Name string `json:"name"`
	}

	type res struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Name      string    `json:"name"`
	}

	var name body
	req, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(req, &name)

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, res{Id: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name.Name})
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbUrl := os.Getenv("CONN")
	db, err := sql.Open("postgres", dbUrl)
	dbQueries := database.New(db)

	_ = apiConfig{DB: dbQueries}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /v1/healthz", healthHandler)
	serveMux.HandleFunc("GET /v1/err", errorHandler)
	serveMux.HandleFunc("POST /v1/users", createUserHandler)

	server := http.Server{Handler: serveMux, Addr: "localhost:" + port}
	fmt.Println("[Info] Starting server on port", 8080)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Info] Server ended")
}
