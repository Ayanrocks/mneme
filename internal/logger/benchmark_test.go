package logger

import (
	"bytes"
	"testing"
)

// BenchmarkZerolog benchmarks the zerolog logger performance
func BenchmarkZerolog(b *testing.B) {
	// Initialize logger with console output
	Init(false, false, false, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("This is a test log message with some data")
	}
}

// BenchmarkZerologWithFields benchmarks zerolog with formatted messages
func BenchmarkZerologWithFields(b *testing.B) {
	Init(false, false, false, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("User action performed by %s", "testuser")
	}
}

// BenchmarkFmtPrintln benchmarks standard fmt.Println for comparison
func BenchmarkFmtPrintln(b *testing.B) {
	// Use a buffer to avoid console output
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.WriteString("This is a test log message with some data\n")
	}
}
