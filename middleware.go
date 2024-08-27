package main

import (
	"github.com/saubuny/bootdev-rss/internal/database"
	"net/http"
	"strings"
)

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
