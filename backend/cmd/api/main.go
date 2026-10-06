package main

import (
	"log"
	"net/http"

	"leave-management/internal/config"
	"leave-management/internal/database"
	"leave-management/internal/httpapi"
	"leave-management/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	dataStore := store.New(db)
	handler := httpapi.New(dataStore, cfg.JWTSecret, cfg.JWTExpiry, cfg.CORSOrigin)

	log.Println("server running on", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, handler))
}
