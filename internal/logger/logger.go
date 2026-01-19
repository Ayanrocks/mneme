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
)

// Init initializes the global logger with CLI-optimized settings
func Init(verbose bool, quiet bool, jsonOutput bool) {
	var output io.Writer = os.Stdout

	// Configure console writer for human-readable output
	if !jsonOutput {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Set log level based on flags
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
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
}

// Get returns the global logger instance
func Get() *Logger {
	if log == nil {
		// Initialize with default settings if not already initialized
		Init(false, false, false)
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
