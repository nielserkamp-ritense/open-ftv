// Package ldv contains the code to interface with Logboek Dataverwerkingen.
package ldv

import (
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/opentelemetry"
)

// New instantiates a new logger for Logboek Dataverwerkingen.
func New(cfg *config.Config, logger *slog.Logger) opentelemetry.Logger {
	l, err := opentelemetry.New(cfg.OpenTelServiceName, cfg.OpenTelURL, cfg.OpenTelTimeout)
	if err != nil && cfg.OpenTelServiceName != "" && cfg.OpenTelURL != "" {
		logger.Error("failed to initialize OpenTelemetry logger for LDV; falling back to STDOUT", "error", err)
		l, err = opentelemetry.New(cfg.OpenTelServiceName, "", cfg.OpenTelTimeout)
	}

	if err != nil || l == nil {
		logger.Error("failed to initialize OpenTelemetry logger for LDV; no more options", "error", err)
		return nil
	}

	logger.Info("OpenTelemetry logger for LDV initialized")
	return l
}
