package main

import (
	"context"
	"flag"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/config"
	"github.com/kilowatt-/ImageRepository/database"
	"github.com/kilowatt-/ImageRepository/routes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const DEFAULTPORT = "3000"
const JWTKeyNotFound = "jwt key not found"

func main() {

	if loadErr := config.InitializeEnvironmentVariables(); loadErr != nil {
		log.Fatal(loadErr)
	}

	if dbConnErr := database.Connect(); dbConnErr != nil {
		log.Fatal(dbConnErr)
	}

	if _, jwtKeyExists := os.LookupEnv("JWT_KEY"); !jwtKeyExists {
		log.Fatal(JWTKeyNotFound)
	}

	corsOrigins, corsExists := os.LookupEnv("ALLOWED_CORS_ORIGINS");

	if !corsExists {
		log.Println("You are initializing the server without any allowed CORS origins. CORS requests will not work!")
	}

	PORT, portExists := os.LookupEnv("GO_PORT")

	if !portExists {
		PORT = DEFAULTPORT
	}

	r := mux.NewRouter()

	routes.RegisterRoutes(r)

	allowedOrigins := handlers.AllowedOrigins([]string{corsOrigins})
	allowedCredentials := handlers.AllowCredentials()
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type, Set-Cookie, *"})

	srv := &http.Server{
		Addr : ":" + PORT,
		Handler: handlers.CORS(allowedOrigins, allowedCredentials, allowedHeaders)(r),
	}

	log.Println("Listening on port " + PORT)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second * 60, "time server waits for other services to finalise")
	flag.Parse()

	signal.Notify(c, os.Interrupt)

	<-c

	if err := database.Disconnect(); err != nil {
		log.Println(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), wait)

	defer cancel()

	_ = srv.Shutdown(ctx)

	log.Println("Shutting down")

	os.Exit(0)
}
