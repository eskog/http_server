package main

import (
	"database/sql"
	"http_server/internal/database"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	cfg := apiConfig{}
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	cfg.Platform = os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Can not connect to database: %s", err)
	}
	cfg.queries = *database.New(db)
	cfg.secret = os.Getenv("SECRET")
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.Handle("GET /admin/metrics", cfg.metrics())
	mux.Handle("GET /api/chirps", cfg.getAllChirps())
	mux.Handle("GET /api/chirps/{chirpID}", cfg.getSingleChirp())
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.Handle("POST /admin/reset", cfg.reset())
	mux.Handle("POST /api/users", cfg.createUser())
	mux.Handle("POST /api/chirps", cfg.postChirp())
	mux.Handle("POST /api/login", cfg.login())
	mux.Handle("POST /api/refresh", http.HandlerFunc(cfg.refreshToken))
	mux.Handle("POST /api/revoke", http.HandlerFunc(cfg.revokeRefreshToken))

	httpserver := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	httpserver.ListenAndServe()
}
