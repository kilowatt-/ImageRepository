package main

import (
	"github.com/kilowatt-/ImageRepository/config"
	"github.com/kilowatt-/ImageRepository/controller"
	"github.com/kilowatt-/ImageRepository/routes"
	"log"
	"net/http"
	"os"
)

const DEFAULTPORT = "3000"

func main() {
	routes.RegisterRoutes()

	if loadErr := config.InitializeEnvironmentVariables(); loadErr != nil {
		panic("No .env file found; exiting")
	}

	_ = controller.ConnectToDB()

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
