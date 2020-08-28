package main

import (
	"github.com/kilowatt-/ImageRepository/config"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/routes"
	"log"
	"net/http"
	"os"
)

const DEFAULTPORT = "3000"

func main() {
	routes.RegisterRoutes()

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

	log.Println("Listening on port " + PORT)

	if err := http.ListenAndServe(":" + PORT, nil); err != nil {
		log.Fatal(err)
	}

	if err := database.Disconnect(); err != nil {
		log.Fatal(err)
	}
}
