package main

import (
	"github.com/Kong/go-pdk/server"
)

// This plugin should be executed after any authentication plugins enabled on the Service or Route.
// The priority is set to execute this plugin after the response-ratelimiting plugin:
// https://docs.konghq.com/2.0.x/plugin-development/custom-logic/#plugins-execution-order
const (
	version  = "1.0.0"
	priority = 899
)

func main() {
	_ = server.StartServer(newConfig, version, priority)
}
