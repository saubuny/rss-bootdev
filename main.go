package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type res struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, res{Status: "OK"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /v1/healthz", healthHandler)
	serveMux.HandleFunc("GET /v1/err", errorHandler)

	server := http.Server{Handler: serveMux, Addr: "localhost:" + port}
	fmt.Println("[Info] Starting server on port", 8080)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Info] Server ended")
}
