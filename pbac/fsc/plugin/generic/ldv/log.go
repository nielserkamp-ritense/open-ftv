// Package ldv contains the code to interface with Logboek Dataverwerkingen.
package ldv

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/opentelemetry"
)

// LDV represents the interface for logging messages to Logboek Dataverwerkingen.
type LDV interface {
	StartSpan(ctx context.Context, attributes ...attribute.KeyValue) (context.Context, trace.Span)
	Shutdown(ctx context.Context) error
}

// New instantiates a new logger for Logboek Dataverwerkingen.
func New(cfg *config.Config, logger *slog.Logger) LDV {
	url, aid := cfg.OpenTelURL, cfg.OpenTelActivityID

	u, err := uuid.Parse(aid)
	if err != nil {
		logger.Error("invalid ActivityID for LDV", "aid", aid, "err2", err)
		return nil
	}

	if url == "" {
		url = "slog" // fall back to slog logger by default.
	}

	ldvCfg := &opentelemetry.LoggerConfig{
		Service:      cfg.OpenTelServiceName,
		URL:          url,
		PrettyPrint:  cfg.OpenTelPretty,
		Logger:       logger,
		BatchTimeout: cfg.OpenTelTimeout,
	}

	var out opentelemetry.Logger
	if out, err = opentelemetry.New(ldvCfg); err != nil && cfg.OpenTelServiceName != "" && !strings.EqualFold(url, "slog") {
		logger.Error("failed to initialize OpenTelemetry logger for LDV; falling back to slog", "url", url, "error", err)
		url = "slog"
		ldvCfg.URL = url
		out, err = opentelemetry.New(ldvCfg)
	}

	if err != nil || out == nil {
		logger.Error("failed to initialize OpenTelemetry logger for LDV; no more options", "service", cfg.OpenTelServiceName, "url", url, "activityID", u.String(), "error", err)
		return nil
	}

	logger.Info("OpenTelemetry logger for LDV initialized", "service", cfg.OpenTelServiceName, "url", url, "activityID", u.String())
	return &ldv{activityID: u.String(), ldv: out}
}

// StartSpan implements the LDV interface.
func (l *ldv) StartSpan(ctx context.Context, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	attributes = append(attributes, attribute.String("authz.activity.id", l.activityID))

	opts := []trace.SpanStartOption{
		trace.WithTimestamp(time.Now().UTC()),
		trace.WithAttributes(attributes...),
	}

	return l.ldv.StartSpan(ctx, l.activityID, opts...)
}

// Shutdown implements the LDV interface.
func (l *ldv) Shutdown(ctx context.Context) error {
	return l.ldv.Shutdown(ctx)
}

type ldv struct {
	activityID string
	ldv        opentelemetry.Logger
}
