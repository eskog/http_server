package main

import (
	"log"
	"net/http"
)

func healthz(rw http.ResponseWriter, r *http.Request) {
	log.Println("Healthz triggered")
	rw.Header().Add("Content-Type:", "text/plain")
	rw.Header().Add("charset", "utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("ok"))
}
