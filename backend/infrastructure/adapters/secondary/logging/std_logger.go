package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// StdLogger implements domain.Logger using Go's standard log.
type StdLogger struct {
	mu sync.Mutex // thread-safe
}

// NewStdLogger creates a new standard logger.
func NewStdLogger() *StdLogger {
	return &StdLogger{}
}

// Info logs an informational message.
func (l *StdLogger) Info(msg string, keysAndValues ...any) {
	l.log("INFO", msg, keysAndValues...)
}

// Warn logs a warning message.
func (l *StdLogger) Warn(msg string, keysAndValues ...any) {
	l.log("WARN", msg, keysAndValues...)
}

// Error logs an error message.
func (l *StdLogger) Error(msg string, keysAndValues ...any) {
	l.log("ERROR", msg, keysAndValues...)
}

// Fatal logs a fatal message and exits with a non-zero status.
// The message is treated as plain text, never as a format string.
func (l *StdLogger) Fatal(msg string, keysAndValues ...any) {
	l.log("FATAL", msg, keysAndValues...)
	os.Exit(1)
}

// log writes a formatted log entry.
// The message is treated as plain text; only the key/value pairs are formatted.
func (l *StdLogger) log(level, msg string, keysAndValues ...any) {
	var sb strings.Builder
	sb.WriteString(level)
	sb.WriteString(": ")
	sb.WriteString(msg)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fmt.Fprintf(&sb, " %v=%v", keysAndValues[i], keysAndValues[i+1])
		} else {
			fmt.Fprintf(&sb, " %v=<missing>", keysAndValues[i])
		}
	}

	log.Print(sb.String())
}

var _ ports.Logger = (*StdLogger)(nil)
