package main

import (
	"encoding/base64"
	"errors"
	"os"
)

const appsPathKey = "APPS_PATH"
const googleApplicationCredentialsKey = "GOOGLE_APPLICATION_CREDENTIALS"

type Config struct {
	googleCredentials string
	AppsPath          string
}

func (c *Config) GetGoogleApplicationCredentials() ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(c.googleCredentials)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func GetConfig() (*Config, error) {
	appsPath := os.Getenv(appsPathKey)
	if appsPath == "" {
		return nil, errors.New("Could not get APPS_PATH variable")
	}
	googleCredentials := os.Getenv(googleApplicationCredentialsKey)

	return &Config{googleCredentials: googleCredentials, AppsPath: appsPath}, nil
}
