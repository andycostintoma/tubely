package main

import (
	"github.com/andycostintoma/tubely/internal/server"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Couldn't load environment variables")
	}

	app, err := server.NewServer()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Server listening on %v", app.Addr)

	err = app.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
