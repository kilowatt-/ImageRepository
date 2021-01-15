package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func InitializeEnvironmentVariables() error {
	if _, productionExists := os.LookupEnv("PRODUCTION"); !productionExists {
		log.Println("Not in production environment; loading local .env file")
		if err := godotenv.Load(); err != nil {
			return err
		}
	}

	return nil
}
