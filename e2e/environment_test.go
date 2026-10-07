//go:build e2e

package e2e

import (
	"os"
	"time"
)

// The defaults point at the local setup of the README. Another environment only needs other values.
type environment struct {
	laadpalenURL     string
	vlierdamUsername string
	vlierdamPassword string
	browserPath      string
	timeout          time.Duration
	tags             string
}

func environmentFromEnv() (environment, error) {
	timeout, err := time.ParseDuration(getEnv("E2E_TIMEOUT", "30s"))
	if err != nil {
		return environment{}, err
	}

	return environment{
		laadpalenURL:     getEnv("E2E_LAADPALEN_URL", "http://localhost:8080"),
		vlierdamUsername: getEnv("E2E_VLIERDAM_BEHEER_USERNAME", "auditor-user"),
		vlierdamPassword: getEnv("E2E_VLIERDAM_BEHEER_PASSWORD", "password"),
		browserPath:      os.Getenv("E2E_BROWSER_PATH"),
		timeout:          timeout,
		tags:             os.Getenv("E2E_TAGS"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
