package config

import (
	"github.com/joho/godotenv"
)

func InitializeEnvironmentVariables() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	return nil
}
