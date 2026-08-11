package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBNAME          string
	DBPASS          string
	DBHOST          string
	DBPORT          string
	DBUSER          string
	PoiskKinoApiKey string
}

func New() (*Config, error) {
	DBNAME, err := getEnv("DB_NAME")
	if err != nil {
		return nil, err
	}
	DBPASS, err := getEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}
	DBHOST, err := getEnv("DB_HOST")
	if err != nil {
		return nil, err
	}
	DBPORT, err := getEnv("DB_PORT")
	if err != nil {
		return nil, err
	}
	DBUSER, err := getEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	PoiskKinoApiKey, err := getEnv("POISK_KINO_API_KEY")
	if err != nil {
		return nil, err
	}

	return &Config{
		DBNAME:          DBNAME,
		DBPASS:          DBPASS,
		DBHOST:          DBHOST,
		DBPORT:          DBPORT,
		DBUSER:          DBUSER,
		PoiskKinoApiKey: PoiskKinoApiKey,
	}, nil
}

func getEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("%s is required", key)
}
