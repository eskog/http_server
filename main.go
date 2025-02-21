package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
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
		cfg.fileServerHits.Swap(0)
		rw.WriteHeader(http.StatusOK)
	})
}

func healthz(rw http.ResponseWriter, r *http.Request) {
	log.Println("Healthz triggered")
	rw.Header().Add("Content-Type:", "text/plain")
	rw.Header().Add("charset", "utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.Handle("GET /admin/metrics", cfg.metrics())
	mux.Handle("POST /admin/reset", cfg.reset())
	mux.HandleFunc("GET /api/healthz", healthz)
	httpserver := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	httpserver.ListenAndServe()
}
