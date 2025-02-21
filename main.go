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
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		response := fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())
		rw.Write([]byte(response))
	})
}

func (cfg *apiConfig) reset() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Swap(0)
		rw.WriteHeader(http.StatusOK)
	})
}

func returnHealth(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type:", "text/plain")
	rw.Header().Add("charset", "utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.Handle("/metrics", cfg.metrics())
	mux.Handle("/reset", cfg.reset())
	mux.HandleFunc("/healthz", returnHealth)
	httpserver := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	httpserver.ListenAndServe()
}
