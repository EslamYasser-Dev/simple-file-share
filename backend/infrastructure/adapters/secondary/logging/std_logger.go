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
	mu     sync.Mutex // thread-safe
	colors bool
}

// NewStdLogger creates a standard logger with ANSI colors when stderr is a TTY.
func NewStdLogger() *StdLogger {
	return &StdLogger{colors: isTTY(os.Stderr)}
}

// NewStdLoggerPlain creates a logger without ANSI colors (tests, pipes).
func NewStdLoggerPlain() *StdLogger {
	return &StdLogger{colors: false}
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
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

func levelColor(level string) string {
	switch level {
	case "INFO":
		return "\033[36m" // cyan
	case "WARN":
		return "\033[33m" // yellow
	case "ERROR", "FATAL":
		return "\033[31m" // red
	default:
		return "\033[0m"
	}
}

func statusColor(code int) string {
	switch {
	case code >= 500:
		return "\033[31m" // red
	case code >= 400:
		return "\033[33m" // yellow
	case code >= 300:
		return "\033[35m" // magenta
	default:
		return "\033[32m" // green
	}
}

// log writes a formatted log entry.
// The message is treated as plain text; only the key/value pairs are formatted.
func (l *StdLogger) log(level, msg string, keysAndValues ...any) {
	var sb strings.Builder

	if l.colors {
		sb.WriteString(levelColor(level))
		sb.WriteString(level)
		sb.WriteString("\033[0m")
	} else {
		sb.WriteString(level)
	}
	sb.WriteString(": ")
	sb.WriteString(msg)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			fmt.Fprintf(&sb, " %v=<missing>", keysAndValues[i])
			continue
		}
		key := fmt.Sprintf("%v", keysAndValues[i])
		val := fmt.Sprintf("%v", keysAndValues[i+1])

		// Color HTTP status codes for quicker scanning.
		if l.colors && key == "status" {
			var code int
			if _, err := fmt.Sscanf(val, "%d", &code); err == nil {
				fmt.Fprintf(&sb, " status=%s%d\033[0m", statusColor(code), code)
				continue
			}
		}

		fmt.Fprintf(&sb, " %v=%v", key, val)
	}

	log.Print(sb.String())
}

var _ ports.Logger = (*StdLogger)(nil)
