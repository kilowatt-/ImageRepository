package main

import (
	"github.com/kilowatt-/ImageRepository/routes"
	"log"
	"net/http"
)

func main() {

	routes.RegisterRoutes()

	log.Println("Listening on port 7000")

	err := http.ListenAndServe(":7000", nil)

	if err != nil {
		log.Fatal(err)
	}
}
