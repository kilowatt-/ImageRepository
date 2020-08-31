package main

import (
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/config"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/routes"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

const DEFAULTPORT = "3000"

func main() {

	r := mux.NewRouter()

	routes.RegisterRoutes(r)

	if loadErr := config.InitializeEnvironmentVariables(); loadErr != nil {
		log.Fatal(loadErr)
	}

	if dbConnErr := database.Connect(); dbConnErr != nil {
		log.Fatal(dbConnErr)
	}

	PORT, portExists := os.LookupEnv("GO_PORT")

	if !portExists {
		PORT = DEFAULTPORT
	}

	corsOrigins, corsExists := os.LookupEnv("ALLOWED_CORS_ORIGINS")

	if !corsExists {
		log.Println("You are initializing the server without any allowed CORS origins. CORS requests will not work!")
	}

	// Initialize CORS with options
	c := cors.New(cors.Options{
		AllowedOrigins: []string{corsOrigins},
		AllowCredentials: true,
		AllowedHeaders: []string{"Content-Type, Set-Cookie, *"},
	})

	log.Println("Listening on port " + PORT)

	if err := http.ListenAndServe(":" + PORT, c.Handler(r)); err != nil {
		log.Fatal(err)
	}

	if err := database.Disconnect(); err != nil {
		log.Fatal(err)
	}
}
