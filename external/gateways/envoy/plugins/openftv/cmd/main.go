package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	c, err := loadConfig(os.Args[1:], os.LookupEnv)
	if errors.Is(err, flag.ErrHelp) {
		return
	}

	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded", "config", c)

	address := fmt.Sprintf("%s:%d", c.Host, c.Port)

	srv, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("failed to initialize AuthZEN plugin service", "address", address, "error", err)
		os.Exit(1)
	}

	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	s := grpc.NewServer()
	auth.RegisterAuthorizationServer(s, newAuthZEN(c, logger))

	logger.Info("AuthZEN plugin service initialized successfully", "address", address)

	go func() {
		defer s.Stop()
		if err = s.Serve(srv); err != nil {
			logger.Error("failed to run AuthZEN plugin service", "address", address, "error", err)
			os.Exit(1)
		}
	}()

	<-intChan

	logger.Info("AuthZEN plugin service shut down successfully", "address", address)
}
