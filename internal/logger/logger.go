package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with CLI-friendly configuration
type Logger struct {
	zerolog.Logger
}

var (
	// Global logger instance
	log *Logger
	// User-friendly logger without timestamps
	userLog *Logger
)

// parseLogLevel converts a log level string to zerolog.Level
// Returns info level as default for empty or invalid values
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		// Default to info for empty or invalid values
		return zerolog.InfoLevel
	}
}

// Init initializes the global logger with CLI-optimized settings
func Init(verbose bool, quiet bool, jsonOutput bool, logLevel string) {
	var output io.Writer = os.Stdout

	// Configure console writer for human-readable output
	if !jsonOutput {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Set log level from config, defaulting to info if invalid/empty
	zerolog.SetGlobalLevel(parseLogLevel(logLevel))

	// CLI flags override config settings
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	// Create logger with zero-allocation optimizations
	logger := zerolog.New(output).
		With().
		Timestamp().
		Logger()

	log = &Logger{logger}

	// Create user-friendly logger without timestamps
	var userOutput io.Writer = os.Stdout
	if !jsonOutput {
		userOutput = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "",
			NoColor:    false,
		}
	}

	userLogger := zerolog.New(userOutput).
		With().
		Logger()

	userLog = &Logger{userLogger}
}

// Get returns the global logger instance
func Get() *Logger {
	if log == nil {
		// Initialize with default settings if not already initialized
		Init(false, false, false, "")
	}
	return log
}

// Debug logs a debug message
func Debug(msg string) {
	Get().Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Get().Debug().Msgf(format, args...)
}

// Info logs an info message
func Info(msg string) {
	Get().Info().Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Get().Info().Msgf(format, args...)
}

// Warn logs a warning message
func Warn(msg string) {
	Get().Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Get().Warn().Msgf(format, args...)
}

// Error logs an error message
func Error(msg string) {
	Get().Error().Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Get().Error().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Get().Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Get().Fatal().Msgf(format, args...)
}

// With creates a logger with context fields
func With() zerolog.Context {
	return Get().With()
}

// WithField creates a logger with a single field
func WithField(key string, value interface{}) zerolog.Context {
	return Get().With().Str(key, value.(string))
}
