package config

import (
	"github.com/joho/godotenv"
	"os"
)

func InitializeEnvironmentVariables() error {
	if _, productionExists := os.LookupEnv("PRODUCTION"); !productionExists {
		if err := godotenv.Load(); err != nil {
			return err
		}
	}

	return nil
}
