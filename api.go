package main

import (
	"encoding/json"
	"fmt"
	"http_server/internal/database"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/google/uuid"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        database.Queries
	Platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1) // Increment counter
		log.Printf("Middleware hit! Counter: %d, Path: %s\n", cfg.fileServerHits.Load(), r.URL.Path)
		rw.Header().Add("Cache-Control", "no-cache")
		next.ServeHTTP(rw, r) // Call the next handler
	})
}

func (cfg *apiConfig) metrics() http.Handler {
	log.Println("Metrics triggered")
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Header().Add("Content-Type", "text/html")
		fmt.Fprintf(rw, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileServerHits.Load())
	})
}

func (cfg *apiConfig) reset() http.Handler {
	log.Println("Reset triggered")
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if cfg.Platform != "dev" {
			http.Error(rw, "Forbidden", http.StatusForbidden)
		}
		cfg.fileServerHits.Swap(0)
		cfg.queries.DropAllUsers(r.Context())
		cfg.queries.DropAllChirps(r.Context())
		rw.WriteHeader(http.StatusOK)
	})
}

func (c *apiConfig) createUser() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		type request struct {
			Email string `json:"email"`
		}
		var req request

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(rw, "Can not decode request", http.StatusUnprocessableEntity)
			return
		}
		createdUser, err := c.queries.CreateUser(r.Context(), req.Email)
		if err != nil {
			http.Error(rw, "Interal database error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(&createdUser)
		rw.Header().Add("Content-Type", "application/json")

	})
}

func (c *apiConfig) getAllChirps() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		data, err := c.queries.GetAllChirps(r.Context())
		if err != nil {
			http.Error(rw, "Can not retrieve data", http.StatusNotFound)
		}
		if err := json.NewEncoder(rw).Encode(&data); err != nil {
			http.Error(rw, "Could not marshal data", http.StatusInternalServerError)
		}
		rw.WriteHeader(http.StatusOK)
	})
}

func (c *apiConfig) getSingleChirp() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			http.Error(rw, "Invalid UUID", http.StatusUnprocessableEntity)
			return
		}
		data, err := c.queries.GetSingleChirp(r.Context(), id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(rw).Encode(&data)
		rw.Header().Add("Content-Type", "applicatio/json")
		rw.WriteHeader(http.StatusOK)

	})
}

func (c *apiConfig) postChirp() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		type requestStruct struct {
			Body    string `json:"body"`
			User_id string `json:"user_id"`
		}
		defer r.Body.Close()
		var req requestStruct
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(rw, "Error unmarshaling request data", http.StatusUnprocessableEntity)
			return
		}
		if len(req.Body) > 140 {
			http.Error(rw, `{"error": "Chirp is too long"}`, http.StatusBadRequest)
			return
		}

		req.Body = cleanupInput(req.Body)
		user_id, _ := uuid.Parse(req.User_id)
		data := database.CreateChirpParams{
			Body:   req.Body,
			UserID: uuid.NullUUID{UUID: user_id, Valid: true},
		}

		createdChirp, err := c.queries.CreateChirp(r.Context(), data)
		if err != nil {
			http.Error(rw, "Interal database error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(&createdChirp)
		rw.Header().Add("Content-Type", "application/json")

	})
}
