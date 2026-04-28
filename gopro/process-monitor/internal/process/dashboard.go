package process

import (
	"context"
	"time"
)

// Dashboard displays process information in a TUI format.
//
// Refs:
// - Go time: https://pkg.go.dev/time
// - Ticker: https://pkg.go.dev/time#NewTicker
// - Context cancellation: https://pkg.go.dev/context
func NewDashboard(ctx context.Context, refreshInterval time.Duration) *Dashboard {
	return &Dashboard{
		ctx:            ctx,
		refreshInterval: refreshInterval,
	}
}

type Dashboard struct {
	ctx            context.Context
	refreshInterval time.Duration
	processes      []Process
	sortColumn    string
	sortAscending bool
}

// Refresh fetches the latest process list from the system.
//
// Implements: Call ListProcesses and update internal state
//
// Refs:
// - ListProcesses: process.go
// - Error handling: https://pkg.go.dev/fmt
func (d *Dashboard) Refresh() error {
	// TODO: Implement Refresh
	//
	// Steps:
	// 1. Check if d.ctx is done: select { case <-d.ctx.Done(): return d.ctx.Err() }
	// 2. Call ListProcesses(d.ctx)
	// 3. Store results in d.processes
	// 4. Return error if any
	//
	// Error handling:
	// - Return fmt.Errorf("refresh failed: %w", err) on ListProcesses failure
	// - Handle context cancellation gracefully
	panic("implement Refresh")
}

// Render returns a string representation of the dashboard.
//
// Implements: Format processes for terminal display
//
// Refs:
// - Go strings: https://pkg.go.dev/strings
// - fmt: https://pkg.go.dev/fmt
//
// Output format (example):
// PID    USER      %CPU  %MEM  RSS     STATE  COMMAND
// 1     root      0.0   0.0   0       S      launchd
// 1234  user      5.2  2.1  1048576 R       chrome
func (d *Dashboard) Render() string {
	// TODO: Implement Render
	//
	// Steps:
	// 1. Build header line
	// 2. Format each process: fmt.Sprintf("%-8d %-8s %5.1f %5.1f %-8s %-8s %s\n", ...)
	// 3. Sort d.processes by d.sortColumn if set
	// 4. Return combined string
	//
	// Column widths:
	// - PID: 8 (right-aligned)
	// - USER: 8
	// - %CPU: 5 (1 decimal)
	// - %MEM: 5 (1 decimal)
	// - RSS: 8 (human readable)
	// - STATE: 8
	// - COMMAND: remaining
	//
	// Notes:
	// - Use FormatBytes for RSS/VSZ conversion
	// - Truncate COMMAND if width > terminal width
	panic("implement Render")
}

// Run starts the dashboard in continuous refresh mode.
//
// Implements: Main TUI loop with ticker
//
// Refs:
// - time.Ticker: https://pkg.go.dev/time#NewTicker
// - select statement: https://pkg.go.dev/tour/gomethods#50
func (d *Dashboard) Run() error {
	// TODO: Implement Run
	//
	// Steps:
	// 1. Create ticker: ticker := time.NewTicker(d.refreshInterval)
	// 2. defer ticker.Stop()
	// 3. Loop:
	//    select {
	//    case <-d.ctx.Done():
	//        return d.ctx.Err()
	//    case <-ticker.C:
	//        d.Refresh()
	//        d.Render() -> print to stdout
	//    }
	// 4. Handle signals (optional): use signal.Notify for SIGINT/SIGTERM
	//
	// Notes:
	// - Clear screen before each render: "\033[2J\033[H"
	// - Use ANSI codes for colors: https://en.wikipedia.org/wiki/ANSI_escape_code
	panic("implement Run")
}

// Stop stops the dashboard and releases resources.
//
// Implements: Cancel context and cleanup
func (d *Dashboard) Stop() {
	// TODO: Implement Stop
	//
	// Steps:
	// 1. Context should be cancelled by caller
	// 2. Any cleanup here if needed
	panic("implement Stop")
}

// SortBy changes the sort column and direction.
//
// Implements: Configure sorting for Render
//
// Columns: "pid", "ppid", "user", "cpu", "mem", "rss", "vsz", "state", "command"
func (d *Dashboard) SortBy(column string, ascending bool) error {
	// TODO: Implement SortBy
	//
	// Steps:
	// 1. Validate column is valid
	// 2. Set d.sortColumn and d.sortAscending
	// 3. Return error if invalid column
	//
	// Error handling:
	// - Return fmt.Errorf("invalid sort column: %s", column)
	panic("implement SortBy")
}

// FilterProcesses filters processes based on a predicate.
//
// Implements: Filter d.processes for display
//
// Refs:
// - Go function types as predicates
//
// Example predicates:
// - func(p Process) bool { return p.CPU > 10.0 }
// - func(p Process) bool { return p.User == "root" }
// - func(p Process) bool { return strings.Contains(p.Command, "chrome") }
func (d *Dashboard) FilterProcesses(predicate func(Process) bool) []Process {
	// TODO: Implement FilterProcesses
	//
	// Steps:
	// 1. result := make([]Process, 0)
	// 2. for _, p := range d.processes {
	//        if predicate(p) { result = append(result, p) }
	//    }
	// 3. return result
	panic("implement FilterProcesses")
}

// TopN returns the top N processes by a column.
//
// Implements: Get top processes sorted by CPU, MEM, etc.
//
// Refs:
// - sort.Slice: https://pkg.go.dev/sort
func (d *Dashboard) TopN(column string, n int) []Process {
	// TODO: Implement TopN
	//
	// Steps:
	// 1. Copy d.processes to temp slice
	// 2. sort.Slice with custom less function based on column
	// 3. Return first n elements
	//
	// Columns:
	// - "cpu": sort by CPU descending
	// - "mem": sort by MEM descending
	// - "rss": sort by RSS descending
	// - "pid": sort by PID ascending
	panic("implement TopN")
}

// GetByPID returns a process by PID.
//
// Implements: Find single process in d.processes
func (d *Dashboard) GetByPID(pid int) *Process {
	// TODO: Implement GetByPID
	//
	// Steps:
	// 1. Linear search: for _, p := range d.processes { if p.PID == pid { return &p } }
	// 2. Return nil if not found
	panic("implement GetByPID")
}