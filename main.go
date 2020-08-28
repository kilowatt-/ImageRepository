package main

import (
	"github.com/joho/godotenv"
	"github.com/kilowatt-/ImageRepository/routes"
	"log"
	"net/http"
	"os"
)

const DEFAULTPORT = "3000"

func main() {
	routes.RegisterRoutes()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found!")
	}

	PORT, portExists := os.LookupEnv("GO_PORT")

	if !portExists {
		PORT = DEFAULTPORT
	}

	log.Println("Listening on port " + PORT)

	err := http.ListenAndServe(":" + PORT, nil)

	if err != nil {
		log.Fatal(err)
	}
}
