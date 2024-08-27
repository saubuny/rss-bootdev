package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerAuth := r.Header.Get("Authorization")
		if headerAuth == "" {
			respondWithError(w, 401, "Authorization header missing")
			return
		}

		apiKey := strings.TrimPrefix(headerAuth, "ApiKey ")
		if apiKey == headerAuth {
			respondWithError(w, 401, "Malformed Token")
			return
		}

		user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, 500, "Error getting user by ApiKey: "+err.Error())
			return
		}

		handler(w, r, user)
	}
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

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Name string `json:"name"`
	}

	type res struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Name      string    `json:"name"`
		ApiKey    string    `json:"api_key"`
	}

	var name body
	req, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(req, &name)

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name.Name,
	})

	if err != nil {
		respondWithError(w, 500, "Error Creating User: "+err.Error())
		return
	}

	respondWithJSON(w, 200, res{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      name.Name,
		ApiKey:    user.ApiKey,
	})
}

// TODO: Use the Auth Middleware for this !!
func (cfg *apiConfig) getUserByApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	headerAuth := r.Header.Get("Authorization")
	if headerAuth == "" {
		respondWithError(w, 401, "Authorization header missing")
		return
	}

	apiKey := strings.TrimPrefix(headerAuth, "ApiKey ")
	if apiKey == headerAuth {
		respondWithError(w, 401, "Malformed Token")
		return
	}

	user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		respondWithError(w, 500, "Error getting user by ApiKey: "+err.Error())
		return
	}

	type res struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Name      string    `json:"name"`
		ApiKey    string    `json:"api_key"`
	}

	respondWithJSON(w, 200, res{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    apiKey,
	})
}

func (cfg *apiConfig) createFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type body struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	var b body
	req, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(req, &b)

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	feed, err := cfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      b.Name,
		Url:       b.Url,
		UserID:    user.ID,
	})

	if err != nil {
		respondWithError(w, 500, "Error Creating Feed: "+err.Error())
		return
	}

	feed_follow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		respondWithError(w, 500, "Error Creating Feed Follow: "+err.Error())
		return
	}

	type res struct {
		Feed       database.Feed       `json:"feed"`
		FeedFollow database.FeedFollow `json:"feed_follow"`
	}

	respondWithJSON(w, 200, res{Feed: feed, FeedFollow: feed_follow})
}

func (cfg *apiConfig) getAllFeedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.GetAllFeeds(r.Context())

	if err != nil {
		respondWithError(w, 500, "Error Getting Feeds: "+err.Error())
		return
	}

	respondWithJSON(w, 200, feeds)
}

func (cfg *apiConfig) createFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type body struct {
		FeedId uuid.UUID `json:"feed_id"`
	}

	var b body
	req, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(req, &b)

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    b.FeedId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		respondWithError(w, 500, "Error Creating Feed Follow: "+err.Error())
		return
	}

	respondWithJSON(w, 200, feedFollow)
}

func (cfg *apiConfig) deleteFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	id, err := uuid.Parse(r.PathValue("feedFollowID"))
	if err != nil {
		respondWithError(w, 500, "Error getting feed follow ID: "+err.Error())
		return
	}

	feed_follow, err := cfg.DB.GetFeedFollowById(r.Context(), id)
	if feed_follow.UserID != user.ID {
		respondWithError(w, 401, "This user does not own the given feed follow")
		return
	}

	err = cfg.DB.DeleteFeedFollow(r.Context(), id)
	if err != nil {
		respondWithError(w, 500, "Error Deleting Feed Follow: "+err.Error())
		return
	}

	w.WriteHeader(200)
}

func (cfg *apiConfig) getFeedFollowsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feed_follows, err := cfg.DB.GetFeedFollowsByUserId(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 500, "Error Getting Feed Follows: "+err.Error())
		return
	}
	respondWithJSON(w, 200, feed_follows)
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

	server := http.Server{Handler: serveMux, Addr: "localhost:" + port}
	fmt.Println("[Info] Starting server on port", 8080)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Info] Server ended")
}
