package display

import (
	"github.com/gluppler/process-monitor/internal/process"
)

// TUI handles terminal display rendering.
//
// Refs:
// - ANSI escape codes: https://en.wikipedia.org/wiki/ANSI_escape_code
// - Terminal colors: https://github.com/tmux/tmux/blob/master/FAQ
// - Go os: https://pkg.go.dev/os
type TUI struct {
	processes []process.Process
	width    int
	height   int
}

// NewTUI creates a new TUI instance.
//
// Refs:
// - Terminal size: https://pkg.go.dev/os#Getwd
func NewTUI() *TUI {
	width, height := 80, 24

	// TODO: Get terminal size properly
	// tty, err := os.Open("/dev/tty")
	// defer tty.Close()
	// Use term.GetSize(tty.Fd())
	return &TUI{
		width:    width,
		height:   height,
	}
}

// Render renders the process list to a string.
//
// Implements: Format process list with ANSI colors
//
// Refs:
// - fmt: https://pkg.go.dev/fmt
// - strings.Builder: https://pkg.go.dev/strings#Builder
//
// ANSI colors:
// - Reset: "\033[0m"
// - Red: "\033[31m"
// - Green: "\033[32m"
// - Yellow: "\033[33m"
// - Blue: "\033[34m"
// - Bold: "\033[1m"
//
// Example output with colors:
// [1mPID    USER      %CPU  %MEM  RSS     STATE  COMMAND[0m
// 1     root      0.0   0.0   0       S      launchd
// [33m1234  user      5.2  2.1   100MB   R      chrome[0m
func (t *TUI) Render(processes []process.Process) string {
	// TODO: Implement Render
	//
	// Steps:
	// 1. Create strings.Builder
	// 2. Write header with bold: fmt.Sprintf("\033[1m%-8s %-10s...\033[0m\n", ...)
	// 3. For each process:
	//    - Color code by state: R=yellow, S=green, Z=red
	//    - Format line: fmt.Sprintf("%-8d %-10s %5.1f %5.1f...\n", ...)
	//    - Write to builder
	// 4. Return builder.String()
	//
	// Colors by state:
	// - "R" (running): Yellow
	// - "S" (sleeping): Green
	// - "Z" (zombie): Red
	// - Others: Default
	panic("implement Render")
}

// SetSize sets the terminal dimensions.
//
// Implements: Update width/height for rendering
func (t *TUI) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// RenderTable renders processes as a formatted table.
//
// Implements: Simple table output without full TUI features
//
// Refs:
// - fmt.Printf: https://pkg.go.dev/fmt
//
// Column formatting:
// - Use negative width for left-align: %-8d
// - Use positive width for right-align: %8d
// - Precision for floats: %.1f
func RenderTable(processes []process.Process) string {
	// TODO: Implement RenderTable
	//
	// Steps:
	// 1. Header: fmt.Sprintf("%-8s %-8s %5s %5s %8s %8s %s\n", "PID", "USER", "%CPU", "MEM", "RSS", "VSZ", "STATE", "COMMAND")
	// 2. Separator: strings.Repeat("-", totalWidth) + "\n"
	// 3. For each process:
	//    - Format RSS with human readable: FormatBytes(p.RSS)
	//    - Format line
	// 4. Return combined string
	//
	// Notes:
	// - Total width ~60 characters
	// - Truncate COMMAND if too long
	panic("implement RenderTable")
}

// ClearScreen sends terminal escape code to clear the screen.
//
// Refs:
// - ANSI escape codes for screen clearing
//
// Code: \033[2J clears screen, \033[H moves cursor to home
func ClearScreen() string {
	return "\033[2J\033[H"
}

// MoveCursor moves cursor to position.
//
// Refs:
// - ANSI escape codes for cursor positioning
//
// Code format: \033[<row>;<col>H
func MoveCursor(row, col int) string {
	// Note: Will need fmt.Sprintf when implemented
	return ""
}

// HideCursor hides the cursor.
//
// Ref: ANSI hide cursor code
func HideCursor() string {
	return "\033[?25l"
}

// ShowCursor shows the cursor.
//
// Ref: ANSI show cursor code
func ShowCursor() string {
	return "\033[?25h"
}

// SaveCursor saves cursor position.
//
// Ref: ANSI save cursor code
func SaveCursor() string {
	return "\033[s"
}

// RestoreCursor restores cursor position.
//
// Ref: ANSI restore cursor code
func RestoreCursor() string {
	return "\033[u"
}

// GetTerminalSize returns the current terminal size.
//
// Refs:
// - sys package: https://pkg.go.dev/syscall
// - TIOCGWINSZ: https://man.openbsd.io/tty_ioctl.4
// - golang.org/x/term: https://pkg.go.dev/golang.org/x/term
func GetTerminalSize() (width int, height int, err error) {
	// TODO: Implement GetTerminalSize
	//
	// Using golang.org/x/term:
	// import "golang.org/x/term"
	// width, height, err := term.GetSize(int(os.Stdout.Fd()))
	// return width, height, err
	//
	// Alternative using syscall:
	// var ws syscall.Winsize
	// _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	// return int(ws.Col), int(ws.Row), nil
	//
	// Note: golang.org/x/term is recommended for cross-platform
	panic("implement GetTerminalSize")
	// return 80, 24, nil
}

// colorForState returns ANSI color code for process state.
//
// Refs:
// - Process states in ps: https://man.openbsd.io/ps.1#PROCESS_STATE
func colorForState(state string) string {
	// TODO: Implement colorForState
	//
	// States and colors:
	// - "R" (running): Yellow "\033[33m"
	// - "S" (sleeping): Green "\033[32m"
	// - "D" (disk wait): Cyan "\033[36m"
	// - "Z" (zombie): Red "\033[31m"
	// - "T" (stopped): Blue "\033[34m"
	// - Default: Reset "\033[0m"
	panic("implement colorForState")
}

// Truncate truncates a string to maxWidth.
//
// Refs:
// - strings.Builder for efficiency
func Truncate(s string, maxWidth int) string {
	// TODO: Implement Truncate
	//
	// Steps:
	// 1. If len(s) <= maxWidth: return s
	// 2. Return s[:maxWidth-3] + "..."
	panic("implement Truncate")
}

// PadLeft pads a string to the left to reach width.
//
// Refs:
// - fmt.Sprintf for padding
func PadLeft(s string, width int) string {
	// TODO: Implement PadLeft
	//
	// Steps:
	// 1. If len(s) >= width: return s
	// 2. Return strings.Repeat(" ", width-len(s)) + s
	panic("implement PadLeft")
}

// PadRight pads a string to the right to reach width.
//
// Refs:
// - fmt.Sprintf for padding
func PadRight(s string, width int) string {
	// TODO: Implement PadRight
	//
	// Steps:
	// 1. If len(s) >= width: return s
	// 2. Return s + strings.Repeat(" ", width-len(s))
	panic("implement PadRight")
}

// ProgressBar renders a simple progress bar.
//
// Refs:
// - ANSI block characters: ▏▎▍▌▋▊▉
func ProgressBar(current, total int, width int) string {
	// TODO: Implement ProgressBar
	//
	// Steps:
	// 1. percent := float64(current) / float64(total)
	// 2. filled := int(float64(width) * percent)
	// 3. filledStr := strings.Repeat("█", filled)
	// 4. emptyStr := strings.Repeat("░", width-filled)
	// 5. Return fmt.Sprintf("[%s%s] %.1f%%", filledStr, emptyStr, percent*100)
	panic("implement ProgressBar")
}