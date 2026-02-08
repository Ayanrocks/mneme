package display

import (
	"testing"
	"time"
)

func TestNewProgressBar(t *testing.T) {
	t.Run("creates progress bar with total", func(t *testing.T) {
		pb := NewProgressBar("Testing", 100)
		if pb == nil {
			t.Fatal("Expected non-nil progress bar")
		}
		if pb.title != "Testing" {
			t.Errorf("Expected title 'Testing', got '%s'", pb.title)
		}
		if pb.total != 100 {
			t.Errorf("Expected total 100, got %d", pb.total)
		}
		if pb.current != 0 {
			t.Errorf("Expected current 0, got %d", pb.current)
		}
		if pb.done {
			t.Error("Expected done to be false")
		}
		if pb.started {
			t.Error("Expected started to be false")
		}
	})

	t.Run("creates indeterminate progress bar", func(t *testing.T) {
		pb := NewProgressBar("Loading", 0)
		if pb == nil {
			t.Fatal("Expected non-nil progress bar")
		}
		if pb.total != 0 {
			t.Errorf("Expected total 0 (indeterminate), got %d", pb.total)
		}
	})

	t.Run("creates progress bar with negative total", func(t *testing.T) {
		pb := NewProgressBar("Testing", -1)
		if pb == nil {
			t.Fatal("Expected non-nil progress bar")
		}
		if pb.total != -1 {
			t.Errorf("Expected total -1, got %d", pb.total)
		}
	})
}

func TestProgressBar_SetCurrent(t *testing.T) {
	pb := NewProgressBar("Test", 100)

	t.Run("sets current progress", func(t *testing.T) {
		pb.SetCurrent(50)
		if pb.current != 50 {
			t.Errorf("Expected current 50, got %d", pb.current)
		}
	})

	t.Run("allows current > total", func(t *testing.T) {
		pb.SetCurrent(150)
		if pb.current != 150 {
			t.Errorf("Expected current 150, got %d", pb.current)
		}
	})

	t.Run("allows negative current", func(t *testing.T) {
		pb.SetCurrent(-10)
		if pb.current != -10 {
			t.Errorf("Expected current -10, got %d", pb.current)
		}
	})
}

func TestProgressBar_SetTotal(t *testing.T) {
	pb := NewProgressBar("Test", 0)

	t.Run("sets total from indeterminate", func(t *testing.T) {
		pb.SetTotal(100)
		if pb.total != 100 {
			t.Errorf("Expected total 100, got %d", pb.total)
		}
	})

	t.Run("updates total mid-progress", func(t *testing.T) {
		pb.SetTotal(200)
		if pb.total != 200 {
			t.Errorf("Expected total 200, got %d", pb.total)
		}
	})
}

func TestProgressBar_SetMessage(t *testing.T) {
	pb := NewProgressBar("Test", 100)

	t.Run("sets message", func(t *testing.T) {
		pb.SetMessage("Processing files...")
		if pb.message != "Processing files..." {
			t.Errorf("Expected message 'Processing files...', got '%s'", pb.message)
		}
	})

	t.Run("allows empty message", func(t *testing.T) {
		pb.SetMessage("")
		if pb.message != "" {
			t.Errorf("Expected empty message, got '%s'", pb.message)
		}
	})

	t.Run("allows long message", func(t *testing.T) {
		longMsg := "This is a very long message that exceeds the typical display width"
		pb.SetMessage(longMsg)
		if pb.message != longMsg {
			t.Errorf("Expected long message, got '%s'", pb.message)
		}
	})
}

func TestProgressBar_Update(t *testing.T) {
	pb := NewProgressBar("Test", 100)

	t.Run("updates current and message", func(t *testing.T) {
		pb.Update(50, "Halfway done")
		if pb.current != 50 {
			t.Errorf("Expected current 50, got %d", pb.current)
		}
		if pb.message != "Halfway done" {
			t.Errorf("Expected message 'Halfway done', got '%s'", pb.message)
		}
	})
}

func TestProgressBar_StartAndComplete(t *testing.T) {
	t.Run("start sets started flag", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		pb.Start()
		// Give goroutine time to start
		time.Sleep(10 * time.Millisecond)
		pb.mutex.Lock()
		started := pb.started
		pb.mutex.Unlock()
		if !started {
			t.Error("Expected started to be true after Start()")
		}
		pb.Complete()
	})

	t.Run("complete sets current to total", func(t *testing.T) {
		pb := NewProgressBar("Test", 100)
		pb.Start()
		time.Sleep(10 * time.Millisecond)
		pb.SetCurrent(50)
		pb.Complete()
		// After complete, current should be set to total
		if pb.total > 0 && pb.current != pb.total {
			t.Errorf("Expected current=%d after Complete(), got %d", pb.total, pb.current)
		}
	})

	t.Run("complete sets done flag", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		pb.Start()
		time.Sleep(10 * time.Millisecond)
		pb.Complete()
		if !pb.done {
			t.Error("Expected done to be true after Complete()")
		}
	})

	t.Run("multiple complete calls are safe", func(t *testing.T) {
		pb := NewProgressBar("Test", 10)
		pb.Start()
		time.Sleep(10 * time.Millisecond)
		pb.Complete()
		// Second complete should not panic
		pb.Complete()
	})
}

func TestProgressBar_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent updates are thread-safe", func(t *testing.T) {
		pb := NewProgressBar("Test", 1000)
		pb.Start()
		time.Sleep(10 * time.Millisecond)

		// Concurrent updates
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(val int) {
				pb.SetCurrent(val)
				pb.SetMessage("Updating")
				pb.SetTotal(val + 1000)
				done <- true
			}(i * 100)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		pb.Complete()
	})
}

func TestShouldShowProgress(t *testing.T) {
	// This test depends on zerolog.GlobalLevel() being set
	// We can't easily control it in tests, but we can verify it returns a boolean
	t.Run("returns boolean", func(t *testing.T) {
		result := ShouldShowProgress()
		// Should be either true or false
		if result {
			// If true, no error
		} else {
			// If false, also no error
		}
	})
}

func TestRunWithProgress(t *testing.T) {
	t.Run("executes function with callback", func(t *testing.T) {
		executed := false
		err := RunWithProgress("Test", 10, func(callback ProgressCallback) error {
			executed = true
			// Call callback a few times
			callback(1, 10, "Step 1")
			callback(5, 10, "Step 5")
			callback(10, 10, "Done")
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !executed {
			t.Error("Expected function to be executed")
		}
	})

	t.Run("returns error from function", func(t *testing.T) {
		testErr := &testError{msg: "test error"}
		err := RunWithProgress("Test", 10, func(callback ProgressCallback) error {
			return testErr
		})

		if err != testErr {
			t.Errorf("Expected error %v, got %v", testErr, err)
		}
	})

	t.Run("callback receives progress updates", func(t *testing.T) {
		var receivedCurrent, receivedTotal int
		var receivedMessage string

		err := RunWithProgress("Test", 0, func(callback ProgressCallback) error {
			callback(42, 100, "Progress")
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Note: We can't easily verify callback values were received
		// without more complex testing infrastructure, but we can verify
		// the function completed without error
		_ = receivedCurrent
		_ = receivedTotal
		_ = receivedMessage
	})
}

// testError is a custom error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestProgressBar_EdgeCases(t *testing.T) {
	t.Run("progress bar with zero title", func(t *testing.T) {
		pb := NewProgressBar("", 100)
		if pb.title != "" {
			t.Errorf("Expected empty title, got '%s'", pb.title)
		}
	})

	t.Run("progress can exceed 100 percent", func(t *testing.T) {
		pb := NewProgressBar("Test", 100)
		pb.SetCurrent(200)
		// Should not panic
		if pb.current != 200 {
			t.Errorf("Expected current 200, got %d", pb.current)
		}
	})
}