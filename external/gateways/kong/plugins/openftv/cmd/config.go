package main

import (
	"context"
	"log/slog"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

func newConfig() any {
	return &Config{pep: pep.New(context.Background(), slog.New(slog.NewJSONHandler(os.Stdout, nil)))}
}

type Config struct {
	PDP     string `json:"pdp"`
	Timeout int    `json:"timeout"`
	// hidden fields
	pep *pep.PEP
}
