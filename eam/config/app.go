// Package config contains the configuration shared by the apps, each part registering its own command-line flags.
package config

// ServerApp contains the configuration variables for a service application.
type ServerApp struct {
	Server
	Log
}
