//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// The defaults point at the local setup of the README. Another environment only needs other values.
type environment struct {
	laadpalenURL     string
	vlierdamUsername string
	vlierdamPassword string
	rvigUsername     string
	rvigPassword     string
	rdwUsername      string
	rdwPassword      string
	browserPath      string
	timeout          time.Duration
	concurrency      int
	tags             string
}

func environmentFromEnv() (environment, error) {
	timeout, err := time.ParseDuration(getEnv("E2E_TIMEOUT", "30s"))
	if err != nil {
		return environment{}, err
	}

	concurrency, err := strconv.Atoi(getEnv("E2E_CONCURRENCY", "4"))
	if err != nil {
		return environment{}, err
	}

	return environment{
		laadpalenURL:     getEnv("E2E_LAADPALEN_URL", "http://localhost:8080"),
		vlierdamUsername: getEnv("E2E_VLIERDAM_BEHEER_USERNAME", "auditor-user"),
		vlierdamPassword: getEnv("E2E_VLIERDAM_BEHEER_PASSWORD", "password"),
		rvigUsername:     getEnv("E2E_RVIG_BEHEER_USERNAME", "auditor-user"),
		rvigPassword:     getEnv("E2E_RVIG_BEHEER_PASSWORD", "password"),
		rdwUsername:      getEnv("E2E_RDW_BEHEER_USERNAME", "auditor-user"),
		rdwPassword:      getEnv("E2E_RDW_BEHEER_PASSWORD", "password"),
		browserPath:      os.Getenv("E2E_BROWSER_PATH"),
		timeout:          timeout,
		concurrency:      concurrency,
		tags:             os.Getenv("E2E_TAGS"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// account returns the account for the management interface of the organization, as the frontpage names it.
func (e *environment) account(organization string) (username, password string, err error) {
	switch organization {
	case "Gemeente Vlierdam":
		return e.vlierdamUsername, e.vlierdamPassword, nil
	case "RvIG":
		return e.rvigUsername, e.rvigPassword, nil
	case "RDW":
		return e.rdwUsername, e.rdwPassword, nil
	default:
		return "", "", fmt.Errorf("no account for the management interface of %s", organization)
	}
}
