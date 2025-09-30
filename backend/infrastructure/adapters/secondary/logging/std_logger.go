package logging

import (
	"log"
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
func (l *StdLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log("INFO", msg, keysAndValues...)
}

// Warn logs a warning message.
func (l *StdLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log("WARN", msg, keysAndValues...)
}

// Error logs an error message.
func (l *StdLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log("ERROR", msg, keysAndValues...)
}

// Fatal logs a fatal message and exits.
func (l *StdLogger) Fatal(msg string, keysAndValues ...interface{}) {
	log.Fatalf("FATAL: "+msg, keysAndValues...)
}

// log writes a formatted log entry.
func (l *StdLogger) log(level, msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(keysAndValues) == 0 {
		log.Printf("%s: %s", level, msg)
		return
	}

	// Simple key=value formatting
	var args []interface{}
	format := "%s: " + msg
	args = append(args, level)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			format += " %v=%v"
			args = append(args, keysAndValues[i], keysAndValues[i+1])
		} else {
			format += " %v=<missing>"
			args = append(args, keysAndValues[i])
		}
	}

	log.Printf(format, args...)
}

var _ ports.Logger = (*StdLogger)(nil)
