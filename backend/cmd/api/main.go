package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/config"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/handlers"
)

func main() {
	cfg := config.Load()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlers.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
