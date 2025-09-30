package ports

// Logger defines the interface for logging messages at different severity levels
type Logger interface {
	// Info logs informational messages
	Info(msg string, keysAndValues ...any)
	// Warn logs warning messages
	Warn(msg string, keysAndValues ...any)
	// Error logs error messages
	Error(msg string, keysAndValues ...any)
	// Fatal logs critical messages and terminates the program
	Fatal(msg string, keysAndValues ...any)
}
