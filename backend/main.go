package main

import (
	"log"
	"net/http"
	"os"

	"github.com/riccardo/freezo/backend/internal/api"
	"github.com/riccardo/freezo/backend/internal/db"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./freezer.db"
	}

	db.InitDB(dbPath)
	defer db.DB.Close()

	r := api.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
