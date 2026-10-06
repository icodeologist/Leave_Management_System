package main

import (
	"log"

	"leave-management/internal/config"
	"leave-management/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("database connected")
}
