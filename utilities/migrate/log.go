package migrate

import (
	"fmt"
	"log/slog"
	"strings"
)

type wrapper struct {
	verbose bool
	logger  *slog.Logger
}

// Printf implements the Logger interface.
func (w *wrapper) Printf(format string, v ...interface{}) {
	w.logger.Info(fmt.Sprintf(strings.TrimSpace(format), v...))
}

// Verbose implements the Logger interface.
// Verbose should return true when verbose logging output is wanted
func (w *wrapper) Verbose() bool {
	return w.verbose
}
