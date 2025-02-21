package main

import (
	"net/http"
)

func returnHealth(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type:", "text/plain")
	rw.Header().Add("charset", "utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", returnHealth)
	httpserver := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	httpserver.ListenAndServe()
}
