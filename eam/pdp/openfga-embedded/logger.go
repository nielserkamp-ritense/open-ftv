package openfga_embedded

import (
	"context"
	"log/slog"

	"github.com/openfga/openfga/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newZapper creates a wrapper around the OpenFGA Logger interface (based on Uber's zap).
func newZapper(logger *slog.Logger) logger.Logger {
	return &zapper{logger: logger}
}

// Debug implements the logger.Logger interface.
func (z *zapper) Debug(msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.Debug("openfga", "msg", msg, "fields", value)
}

// Info implements the logger.Logger interface.
func (z *zapper) Info(msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.Info("openfga", "msg", msg, "fields", value)
}

// Warn implements the logger.Logger interface.
func (z *zapper) Warn(msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.Warn("openfga", "msg", msg, "fields", value)
}

// Error implements the logger.Logger interface.
func (z *zapper) Error(msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.Error("openfga", "msg", msg, "fields", value)
}

// Panic implements the logger.Logger interface.
func (z *zapper) Panic(msg string, fields ...zap.Field) {
	// we don't panic for OpenFGA.
	z.Error(msg, fields...)
}

// Fatal implements the logger.Logger interface.
func (z *zapper) Fatal(msg string, fields ...zap.Field) {
	// we don't exit for OpenFGA.
	z.Error(msg, fields...)
}

// With implements the logger.Logger interface.
func (z *zapper) With(fields ...zap.Field) logger.Logger {
	return &zapper{logger: z.logger, fields: append(z.fields, fields...)}
}

// DebugWithContext implements the logger.Logger interface.
func (z *zapper) DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.DebugContext(ctx, "openfga", "msg", msg, "fields", value)
}

// InfoWithContext implements the logger.Logger interface.
func (z *zapper) InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.InfoContext(ctx, "openfga", "msg", msg, "fields", value)
}

// WarnWithContext implements the logger.Logger interface.
func (z *zapper) WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.WarnContext(ctx, "openfga", "msg", msg, "fields", value)
}

// ErrorWithContext implements the logger.Logger interface.
func (z *zapper) ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	value := convertFields(append(z.fields, fields...))
	z.logger.ErrorContext(ctx, "openfga", "msg", msg, "fields", value)
}

// PanicWithContext implements the logger.Logger interface.
func (z *zapper) PanicWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	// we don't panic for OpenFGA.
	z.ErrorWithContext(ctx, msg, fields...)
}

// FatalWithContext implements the logger.Logger interface.
func (z *zapper) FatalWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	// we don't exit for OpenFGA.
	z.ErrorWithContext(ctx, msg, fields...)
}

type zapper struct {
	logger *slog.Logger
	fields []zap.Field
}

func convertFields(fields []zap.Field) map[string]any {
	enc := zapcore.NewMapObjectEncoder()
	for i := range fields {
		fields[i].AddTo(enc)
	}
	return enc.Fields
}
