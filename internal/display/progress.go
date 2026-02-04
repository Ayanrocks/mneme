package display

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	progressWidth = 40          // Width of the progress bar
	clearLine     = "\033[2K"   // ANSI: clear entire line
	moveToStart   = "\r"        // Move cursor to start of line
	saveCursor    = "\033[s"    // ANSI: save cursor position
	restoreCursor = "\033[u"    // ANSI: restore cursor position
	hideCursor    = "\033[?25l" // ANSI: hide cursor
	showCursor    = "\033[?25h" // ANSI: show cursor
)

// Spinner frames for indeterminate progress
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ProgressBar represents a single-line bottom progress bar like apt-get
type ProgressBar struct {
	title      string
	total      int
	current    int
	message    string
	mutex      sync.Mutex
	done       bool
	startTime  time.Time
	spinnerIdx int
	stopChan   chan struct{}
	doneChan   chan struct{}
	started    bool
}

// NewProgressBar creates a new progress bar
// If total is 0 or negative, it will show a spinner (indeterminate mode)
// If total > 0, it will show a progress bar with percentage
func NewProgressBar(title string, total int) *ProgressBar {
	return &ProgressBar{
		title:     title,
		total:     total,
		current:   0,
		message:   "",
		done:      false,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
		doneChan:  make(chan struct{}),
		started:   false,
	}
}

// ShouldShowProgress returns true if the log level is set to "info"
// This determines whether to show the progress bar or fall back to regular logging
func ShouldShowProgress() bool {
	return zerolog.GlobalLevel() == zerolog.InfoLevel
}

// Start begins the progress bar display
func (pb *ProgressBar) Start() {
	pb.mutex.Lock()
	pb.startTime = time.Now()
	pb.started = true
	pb.mutex.Unlock()

	// Hide cursor for cleaner display
	fmt.Print(hideCursor)

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		defer close(pb.doneChan)

		for {
			select {
			case <-pb.stopChan:
				return
			case <-ticker.C:
				pb.render()
			}
		}
	}()
}

// SetCurrent sets the current progress value
func (pb *ProgressBar) SetCurrent(current int) {
	pb.mutex.Lock()
	pb.current = current
	pb.mutex.Unlock()
}

// SetTotal sets the total value (switches to determinate mode if > 0)
func (pb *ProgressBar) SetTotal(total int) {
	pb.mutex.Lock()
	pb.total = total
	pb.mutex.Unlock()
}

// SetMessage sets the current status message
func (pb *ProgressBar) SetMessage(message string) {
	pb.mutex.Lock()
	pb.message = message
	pb.mutex.Unlock()
}

// Update updates both current progress and message
func (pb *ProgressBar) Update(current int, message string) {
	pb.mutex.Lock()
	pb.current = current
	pb.message = message
	pb.mutex.Unlock()
}

// Complete marks the progress as done and cleans up
func (pb *ProgressBar) Complete() {
	pb.mutex.Lock()
	if pb.done {
		pb.mutex.Unlock()
		return
	}
	pb.done = true
	if pb.total > 0 {
		pb.current = pb.total
	}
	pb.mutex.Unlock()

	close(pb.stopChan)
	<-pb.doneChan

	// Clear the progress line and show cursor
	fmt.Print(clearLine + moveToStart + showCursor)
}

// render draws the single-line progress bar at current cursor position
func (pb *ProgressBar) render() {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	if pb.done {
		return
	}

	var line string
	elapsed := time.Since(pb.startTime).Round(time.Second)

	if pb.total > 0 {
		// Determinate mode - show progress bar
		percent := float64(pb.current) / float64(pb.total)
		if percent > 1.0 {
			percent = 1.0
		}

		filled := int(percent * float64(progressWidth))
		if filled > progressWidth {
			filled = progressWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", progressWidth-filled)

		// Format: Title [████████░░░░] 45% (1234/5678) - Message (0m30s)
		msg := pb.message
		if len(msg) > 30 {
			msg = msg[:27] + "..."
		}
		if msg != "" {
			line = fmt.Sprintf("%s [%s] %3.0f%% (%d/%d) - %s (%s)",
				pb.title, bar, percent*100, pb.current, pb.total, msg, elapsed)
		} else {
			line = fmt.Sprintf("%s [%s] %3.0f%% (%d/%d) (%s)",
				pb.title, bar, percent*100, pb.current, pb.total, elapsed)
		}
	} else {
		// Indeterminate mode - show spinner
		frame := spinnerFrames[pb.spinnerIdx%len(spinnerFrames)]
		pb.spinnerIdx++

		msg := pb.message
		if len(msg) > 40 {
			msg = msg[:37] + "..."
		}
		if pb.current > 0 {
			if msg != "" {
				line = fmt.Sprintf("%s %s %d files - %s (%s)", frame, pb.title, pb.current, msg, elapsed)
			} else {
				line = fmt.Sprintf("%s %s %d files (%s)", frame, pb.title, pb.current, elapsed)
			}
		} else {
			if msg != "" {
				line = fmt.Sprintf("%s %s - %s (%s)", frame, pb.title, msg, elapsed)
			} else {
				line = fmt.Sprintf("%s %s (%s)", frame, pb.title, elapsed)
			}
		}
	}

	// Clear line and print new content (stays on same line)
	fmt.Print(clearLine + moveToStart + line)
}

// ProgressCallback is a function type for progress updates
// Used to decouple the index builder from the display package
type ProgressCallback func(current, total int, message string)

// RunWithProgress executes a function with a progress bar if log level is "info"
// Otherwise, it just runs the function with a simple callback that logs normally
func RunWithProgress(title string, total int, fn func(callback ProgressCallback) error) error {
	if !ShouldShowProgress() {
		// Fallback to no-op callback - logs will be handled by normal logger
		return fn(func(current, total int, message string) {
			// No-op
		})
	}

	pb := NewProgressBar(title, total)
	pb.Start()

	callback := func(current, total int, message string) {
		if total > 0 {
			pb.SetTotal(total)
		}
		pb.SetCurrent(current)
		pb.SetMessage(message)
	}

	err := fn(callback)
	pb.Complete()
	return err
}
