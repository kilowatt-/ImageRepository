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
const AWSKeyNotFound = "AWS key not found"
const AWSSecretKeyNotFound = "AWS secret key not found"

func main() {

	if loadErr := config.InitializeEnvironmentVariables(); loadErr != nil {
		log.Fatal(loadErr)
	}

	if dbConnErr := database.Connect(); dbConnErr != nil {
		log.Fatal(dbConnErr)
	}

	defer database.Disconnect()

	if _, jwtKeyExists := os.LookupEnv("JWT_KEY"); !jwtKeyExists {
		log.Fatal(JWTKeyNotFound)
	}

	if _, awsKeyExists := os.LookupEnv("AWS_ACCESS_KEY_ID"); !awsKeyExists {
		log.Fatal(AWSKeyNotFound)
	}

	if _, awsSecretKeyExists := os.LookupEnv("AWS_SECRET_ACCESS_KEY"); !awsSecretKeyExists {
		log.Fatal(AWSSecretKeyNotFound)
	}

	if _, awsRegionExists := os.LookupEnv("AWS_REGION"); !awsRegionExists {
		log.Println("AWS region not found; setting default to us-west-2")
		_ = os.Setenv("AWS_REGION", "us-west-2")
	}

	corsOrigins, corsExists := os.LookupEnv("ALLOWED_CORS_ORIGINS")

	if !corsExists {
		log.Println("You are initializing the server without any allowed CORS origins. CORS requests will not work!")
	}

	PORT, portExists := os.LookupEnv("PORT")

	if !portExists {
		PORT = DEFAULTPORT
	}

	r := mux.NewRouter()

	routes.RegisterRoutes(r)

	allowedOrigins := handlers.AllowedOrigins([]string{corsOrigins})
	allowedCredentials := handlers.AllowCredentials()
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type, Set-Cookie, *"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	srv := &http.Server{
		Addr : ":" + PORT,
		Handler: handlers.CORS(allowedOrigins, allowedCredentials, allowedHeaders, allowedMethods)(r),
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

	ctx, cancel := context.WithTimeout(context.Background(), wait)

	defer cancel()

	_ = srv.Shutdown(ctx)

	log.Println("Shutting down")

	os.Exit(0)
}
