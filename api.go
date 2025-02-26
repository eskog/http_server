package main

import (
	"encoding/json"
	"fmt"
	"http_server/internal/auth"
	"http_server/internal/database"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        database.Queries
	Platform       string
	secret         string
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
		cfg.queries.DropAllTokens(r.Context())
		rw.WriteHeader(http.StatusOK)
	})
}

func (c *apiConfig) createUser() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		type request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		var req request

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(rw, "Can not decode request", http.StatusUnprocessableEntity)
			return
		}
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(rw, "Error creating user password", http.StatusInternalServerError)
			log.Printf("Error creating password hash: %s", err)
			return
		}
		params := database.CreateUserParams{
			Email:          req.Email,
			HashedPassword: hashedPassword,
		}
		createdUser, err := c.queries.CreateUser(r.Context(), params)
		if err != nil {
			http.Error(rw, "Interal database error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		createdUser.HashedPassword = ""
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(&createdUser)
		rw.Header().Add("Content-Type", "application/json")

	})
}

func (c *apiConfig) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type requestStruct struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		type responseStruct struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			Refreshtoken string    `json:"refresh_token"`
		}

		defer r.Body.Close()
		var req requestStruct
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Error decoding request", http.StatusUnprocessableEntity)
			log.Printf("Error decoding login request: %s", err)
			return
		}

		user, err := c.queries.GetUserFromEmail(r.Context(), req.Email)
		if err != nil {
			http.Error(w, "Could not retrieve user", http.StatusInternalServerError)
			log.Printf("Error retrieving user from database: %s", err)
			return
		}
		if err := auth.CheckPasswordHash(req.Password, user.HashedPassword); err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
		res := responseStruct{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        "",
			Refreshtoken: "",
		}

		token, err := auth.MakeJWT(res.ID, c.secret, time.Second*3600)
		if err != nil {
			http.Error(w, "unable to create JWT token", http.StatusInternalServerError)
			log.Printf("Unable to create JWT token: %s", err)
			return
		}
		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			http.Error(w, "Error creating refreshtoken", http.StatusInternalServerError)
			return
		}
		_, err = c.queries.InsertRefreshToken(r.Context(), database.InsertRefreshTokenParams{
			UserID:    res.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().Add(time.Hour * 1440)}) //60 days expire time, 24 hours * 60 days = 1440hours

		if err != nil {
			http.Error(w, "Error inserting token to database", http.StatusInternalServerError)
		}
		res.Refreshtoken = refreshToken
		res.Token = token
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

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
			http.Error(rw, err.Error(), http.StatusNotFound)
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
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(rw, "Could not extract token from headers. Did you send it?", http.StatusUnauthorized)
			return
		}
		log.Printf("Have extracted token: %s", token)
		user_id, err := auth.ValidateJWT(token, c.secret)
		if err != nil {
			http.Error(rw, "Unable to verify token", http.StatusUnauthorized)
			log.Printf("Unable to verify token, err: %s", err)
			return
		}
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
		data := database.CreateChirpParams{
			Body:   req.Body,
			UserID: user_id,
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

func (c *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	type responseStruct struct {
		Token string `json:"token"`
	}
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unable to extract token", http.StatusUnprocessableEntity)
		return
	}
	log.Printf("Token extracted: %s\n", tokenString)
	//TODO: should be refactored at later stage. Not needed to extract entire token object. I just need the UUID.
	token, err := c.queries.GetOneRefreshToken(r.Context(), tokenString)
	if err != nil {
		http.Error(w, "Unable to find refresh token", http.StatusNotFound)
		log.Printf("Error extracting token from psql: %s\n", err)
		return
	}
	log.Printf("Token after sql extraction: %s\n", token.Token)

	if time.Now().After(token.ExpiresAt) || token.RevokedAt.Valid {
		http.Error(w, "Token has expired or has been revoked", http.StatusUnauthorized)
		return
	}
	authToken, err := auth.MakeJWT(token.UserID, c.secret, time.Duration(time.Hour))
	if err != nil {
		http.Error(w, "unable to create new token", http.StatusInternalServerError)
		return
	}
	res := responseStruct{Token: authToken}
	json.NewEncoder(w).Encode(&res)
	w.WriteHeader(http.StatusOK)

}

func (c *apiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unable to find token", http.StatusNotFound)
		log.Printf("Unable to extract header: %s\n", err)
		return
	}
	token, err := c.queries.GetOneRefreshToken(r.Context(), tokenString)
	if err != nil {
		http.Error(w, "Unable to find token", http.StatusNotFound)
		return
	}

	if time.Now().After(token.ExpiresAt) || token.RevokedAt.Valid {
		http.Error(w, "Token has expired or has been revoked", http.StatusUnauthorized)
		return
	}
	if err = c.queries.RevokeToken(r.Context(), tokenString); err != nil {
		http.Error(w, "Could not revoke token", http.StatusInternalServerError)
		log.Printf("Error sending sql: %s\n", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *apiConfig) updateUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unable to extract JWT", http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(tokenString, c.secret)
	type requestStruct struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err != nil {
		http.Error(w, "Token not valid, GTFO", http.StatusUnauthorized)
		return
	}
	var req requestStruct
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "unable to process request", http.StatusUnprocessableEntity)
		return
	}
	req.Password, err = auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Could not process password", http.StatusInternalServerError)
		return
	}
	dbUser, err := c.queries.UpdateUserEmailPassword(r.Context(), database.UpdateUserEmailPasswordParams{
		ID:             userID,
		Email:          req.Email,
		HashedPassword: req.Password,
	})
	if err != nil {
		http.Error(w, "Unable to update database", http.StatusInternalServerError)
		return
	}
	dbUser.HashedPassword = ""
	json.NewEncoder(w).Encode(&dbUser)
	w.WriteHeader(http.StatusOK)

}

func (c *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	/*
		Authenticate user
		Make sure user "owns" the chirp. Then delete chirp
		if successfully deleted. return 204, else return 404
		id, err := uuid.Parse(r.PathValue("chirpID"))
	*/

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unable to extract token", http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(tokenString, c.secret)
	if err != nil {
		http.Error(w, "Token not valid", http.StatusUnauthorized)
		return
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		http.Error(w, "unable to parse chirpID", http.StatusNotFound)
		return
	}
	chirp, err := c.queries.GetSingleChirp(r.Context(), chirpID)
	if err != nil {
		http.Error(w, "Can not find chirp", http.StatusNotFound)
		return
	}
	if chirp.UserID != userID {
		http.Error(w, "Not authorized to delete this chirp", http.StatusForbidden)
		return
	}
	err = c.queries.DeleteSingleChirp(r.Context(), chirp.ID)
	if err != nil {
		http.Error(w, "Chirp not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
