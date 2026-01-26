package logger

import (
	"bytes"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("default initialization", func(t *testing.T) {
		// Reset logger
		log = nil
		userLog = nil

		Init(false, false, false)

		require.NotNil(t, log)
		require.NotNil(t, userLog)
	})

	t.Run("verbose mode", func(t *testing.T) {
		log = nil
		userLog = nil

		Init(true, false, false)

		require.NotNil(t, log)
		require.NotNil(t, userLog)
	})

	t.Run("quiet mode", func(t *testing.T) {
		log = nil
		userLog = nil

		Init(false, true, false)

		require.NotNil(t, log)
		require.NotNil(t, userLog)
	})

	t.Run("json output mode", func(t *testing.T) {
		log = nil
		userLog = nil

		Init(false, false, true)

		require.NotNil(t, log)
		require.NotNil(t, userLog)
	})
}

func TestGet(t *testing.T) {
	t.Run("returns initialized logger", func(t *testing.T) {
		log = nil
		userLog = nil

		logger := Get()
		require.NotNil(t, logger)
	})

	t.Run("initializes if not already initialized", func(t *testing.T) {
		log = nil
		userLog = nil

		logger := Get()
		require.NotNil(t, logger)

		// Second call should return the same instance
		logger2 := Get()
		assert.Equal(t, logger, logger2)
	})
}

func TestLoggingLevels(t *testing.T) {
	// Capture output for testing
	var buf bytes.Buffer
	consoleWriter := zerolog.ConsoleWriter{
		Out:        &buf,
		TimeFormat: "",
		NoColor:    true,
	}

	log = &Logger{
		Logger: zerolog.New(consoleWriter).With().Logger(),
	}

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		// Set log level to Debug for this test
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		Debug("debug message")
		output := buf.String()
		assert.Contains(t, output, "debug message")
		// Reset to Info level for other tests
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	})

	t.Run("Debugf", func(t *testing.T) {
		buf.Reset()
		// Set log level to Debug for this test
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		Debugf("debug %s", "formatted")
		output := buf.String()
		assert.Contains(t, output, "debug formatted")
		// Reset to Info level for other tests
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	})

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		Info("info message")
		output := buf.String()
		assert.Contains(t, output, "info message")
	})

	t.Run("Infof", func(t *testing.T) {
		buf.Reset()
		Infof("info %s", "formatted")
		output := buf.String()
		assert.Contains(t, output, "info formatted")
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		Warn("warn message")
		output := buf.String()
		assert.Contains(t, output, "warn message")
	})

	t.Run("Warnf", func(t *testing.T) {
		buf.Reset()
		Warnf("warn %s", "formatted")
		output := buf.String()
		assert.Contains(t, output, "warn formatted")
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		Error("error message")
		output := buf.String()
		assert.Contains(t, output, "error message")
	})

	t.Run("Errorf", func(t *testing.T) {
		buf.Reset()
		Errorf("error %s", "formatted")
		output := buf.String()
		assert.Contains(t, output, "error formatted")
	})
}

func TestWith(t *testing.T) {
	log = nil
	userLog = nil
	Init(false, false, false)

	t.Run("returns context", func(t *testing.T) {
		ctx := With()
		assert.NotNil(t, ctx)
	})
}

func TestWithField(t *testing.T) {
	log = nil
	userLog = nil
	Init(false, false, false)

	t.Run("returns context with field", func(t *testing.T) {
		ctx := WithField("key", "value")
		assert.NotNil(t, ctx)
	})
}

func TestUserLog(t *testing.T) {
	t.Run("user log without timestamps", func(t *testing.T) {
		var buf bytes.Buffer
		consoleWriter := zerolog.ConsoleWriter{
			Out:        &buf,
			TimeFormat: "",
			NoColor:    true,
		}

		userLog = &Logger{
			Logger: zerolog.New(consoleWriter).With().Logger(),
		}

		// User log should not contain timestamps
		Info("test message")
		output := buf.String()
		// The user log should not have timestamp format
		assert.NotContains(t, output, "T")
	})
}

// Output function tests
func TestSetColors(t *testing.T) {
	t.Run("enable colors", func(t *testing.T) {
		SetColors(true)
		assert.True(t, colorsEnabled)
	})

	t.Run("disable colors", func(t *testing.T) {
		SetColors(false)
		assert.False(t, colorsEnabled)
	})
}

func TestSuccess(t *testing.T) {
	t.Run("prints success message", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		Success("test success message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "test success message")
	})
}

func TestPrint(t *testing.T) {
	t.Run("prints info message", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		Print("test info message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "ℹ")
		assert.Contains(t, output, "test info message")
	})
}

func TestPrintRaw(t *testing.T) {
	t.Run("prints raw message", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintRaw("test raw message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "test raw message")
		assert.NotContains(t, output, "ℹ")
	})
}

func TestWarning(t *testing.T) {
	t.Run("prints warning message", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		Warning("test warning message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "⚠")
		assert.Contains(t, output, "test warning message")
	})
}

func TestPrintError(t *testing.T) {
	t.Run("prints error message to stderr", func(t *testing.T) {
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		SetColors(false)
		PrintError("test error message")

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "✖")
		assert.Contains(t, output, "test error message")
	})
}

func TestHeader(t *testing.T) {
	t.Run("prints header", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		Header("test header")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "test header")
	})
}

func TestSubHeader(t *testing.T) {
	t.Run("prints sub-header", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		SubHeader("test sub-header")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "test sub-header")
	})
}

func TestBullet(t *testing.T) {
	t.Run("prints bullet point", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		Bullet("test bullet")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "•")
		assert.Contains(t, output, "test bullet")
	})
}

func TestList(t *testing.T) {
	t.Run("prints numbered list", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		List([]string{"item1", "item2", "item3"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "1. item1")
		assert.Contains(t, output, "2. item2")
		assert.Contains(t, output, "3. item3")
	})
}

func TestKeyValue(t *testing.T) {
	t.Run("prints key-value pair", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		KeyValue("key", "value")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "key:")
		assert.Contains(t, output, "value")
	})
}

func TestSeparator(t *testing.T) {
	t.Run("prints separator line", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		Separator()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "─")
	})
}

func TestBlank(t *testing.T) {
	t.Run("prints blank line", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		Blank()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "\n")
	})
}

func TestProgress(t *testing.T) {
	t.Run("prints progress message", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		SetColors(false)
		Progress("test progress")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "test progress")
	})
}

func TestClearProgress(t *testing.T) {
	t.Run("clears progress line", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ClearProgress()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "\r\033[K")
	})
}

func TestSpinner(t *testing.T) {
	t.Run("creates new spinner", func(t *testing.T) {
		spinner := NewSpinner()
		assert.NotNil(t, spinner)
		assert.False(t, spinner.active)
	})

	t.Run("starts and stops spinner", func(t *testing.T) {
		spinner := NewSpinner()
		spinner.Start("loading")
		assert.True(t, spinner.active)
		spinner.Stop()
		assert.False(t, spinner.active)
	})
}
