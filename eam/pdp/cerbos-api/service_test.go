package cerbos_api

import "os"

const (
	adminUser = "cerbos"
	adminPswd = "cerbos"
)

func getAddress() string {
	if s := os.Getenv("CERBOS_ADDRESS"); s != "" {
		return s
	}
	return "dns:localhost:6693"
}
