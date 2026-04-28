package process

import (
	"context"
)

// Process represents a running process on the system.
// Fields map to ps command output columns.
//
// Ref: ps man page https://man.openbsd.io/ps.1
type Process struct {
	PID     int     // Process ID
	PPID    int    // Parent Process ID
	User    string // Username running the process
	CPU     float64 // Percentage of CPU used
	MEM     float64 // Percentage of memory used
	RSS     uint64 // Resident Set Size (bytes) - physical memory
	VSZ     uint64 // Virtual Set Size (bytes) - virtual memory
	State   string // Process state: R=running, S=sleeping, Z=zombie, T=stopped
	Command string // Command name (executable)
	Path   string // Full path to executable (not implemented yet)
	Ports  []int // Open network ports (not implemented yet)
	FDs    []string // Open file descriptors (not implemented yet)
}

// Connection represents an active network connection.
//
// Ref: lsof man page https://man.openbsd.io/lsof.8
type Connection struct {
	Protocol    string // TCP, UDP
	LocalIP    string // Local IP address
	LocalPort  int    // Local port number
	ForeignIP string // Foreign IP address
	ForeignPort int    // Foreign port number
	State     string // Connection state: ESTABLISHED, LISTEN, TIME_WAIT, etc.
}

// ListProcesses returns all running processes on the system.
//
// Implements: Parse output from ps command into []Process
//
// Command: ps -ax -o pid,ppid,user,%cpu,%mem,rss,vsz,state,comm= --no-headers
//
// Refs:
// - ps command: https://man.openbsd.io/ps.1
// - Process states: https://man.openbsd.io/ps.1#PROCESS_STATE
// - Go exec: https://pkg.go.dev/os/exec
// - Go strings: https://pkg.go.dev/strings
//
// Example output:
//  1     0 root     0.0  0.0     0     0 S     launchd
//  2     0 root     0.0  0.0     0     0 S     taskgated
func ListProcesses(ctx context.Context) ([]Process, error) {
	// TODO: Implement ListProcesses
	//
	// Steps:
	// 1. Run: exec.CommandContext(ctx, "ps", "-ax", "-o", "pid,ppid,user,%cpu,%mem,rss,vsz,state,comm=", "--no-headers")
	// 2. Capture stdout
	// 3. Split into lines (strings.Split out, "\n")
	// 4. For each line:
	//    a. Split by whitespace (strings.Fields)
	//    b. Parse each field into Process struct fields
	//    c. Handle errors: strconv.Atoi for int fields, strconv.ParseFloat for CPU/MEM
	// 5. Return []Process, nil
	//
	// Error handling:
	// - Return nil, fmt.Errorf("ps failed: %w", err) on command failure
	// - Return partial results with error if some lines fail to parse
	// - Skip header line if present (strings.HasPrefix(line, "PID"))
	panic("implement ListProcesses")
}

// GetProcessInfo returns detailed information for a specific process ID.
//
// Implements: Get full details for a single PID including path
//
// Command: ps -o pid,ppid,user,%cpu,%mem,rss,vsz,state,comm,args= -p <pid>
//
// Refs:
// - ps with -p flag: https://man.openbsd.io/ps.1
//
// Example: Get full path with `ps -o comm= -p <pid>`
func GetProcessInfo(ctx context.Context, pid int) (*Process, error) {
	// TODO: Implement GetProcessInfo
	//
	// Steps:
	// 1. Validate pid > 0
	// 2. Run: ps -o pid,ppid,user,%cpu,%mem,rss,vsz,state,comm,args= -p <pid>
	// 3. Parse output similar to ListProcesses
	// 4. Return single Process, nil
	//
	// Error handling:
	// - Return nil, fmt.Errorf("process %d not found", pid) if ps returns empty
	// - Return nil, fmt.Errorf("invalid pid: %w", err) on validation failure
	panic("implement GetProcessInfo")
}

// GetOpenFDs returns open file descriptors for a process.
//
// Implements: Use lsof to get open files for a PID
//
// Command: lsof -p <pid> -F (or lsof -i -n -P for network only)
//
// Refs:
// - lsof command: https://man.openbsd.io/lsof.8
// - lsof field list: https://man.openbsd.io/lsof.8#OUTPUT
//
// Example output:
// COMMAND PID     USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
// bash    1234    user    0    REG                1,2    0      /dev/null
// bash    1234    user    1    REG                1,2    0      /tmp/foo
func GetOpenFDs(ctx context.Context, pid int) ([]string, error) {
	// TODO: Implement GetOpenFDs
	//
	// Steps:
	// 1. Run: exec.CommandContext(ctx, "lsof", "-p", fmt.Sprintf("%d", pid))
	// 2. Parse output lines
	// 3. Extract file paths from NAME column
	// 4. Return []string, nil
	//
	// Error handling:
	// - Return nil, fmt.Errorf("lsof failed: %w", err) on command failure
	// - Silently return empty slice if process terminated
	panic("implement GetOpenFDs")
}

// GetConnections returns active network connections for a process.
//
// Implements: Use lsof to get network connections for a PID
//
// Command: lsof -i -n -P (filtered by PID)
//
// Refs:
// - lsof -i for network: https://man.openbsd.io/lsof.8#NETWORK_FILES
// - Network states: https://man.openbsd.io/netstat.1
//
// Example output:
// COMMAND PID     USER   PROTOCOL  LOCAL ADDR        FOREIGN ADDR        STATE
// Chrome   1234    user   TCP       192.168.1.5:443    172.217.14.206:443   ESTABLISHED
// Chrome   1234    user   TCP       192.168.1.5:80    172.217.14.206:80    TIME_WAIT
func GetConnections(ctx context.Context, pid int) ([]Connection, error) {
	// TODO: Implement GetConnections
	//
	// Steps:
	// 1. Run: exec.CommandContext(ctx, "lsof", "-i", "-n", "-P", "-p", fmt.Sprintf("%d", pid))
	// 2. Parse output:
	//    - Split line by whitespace
	//    - Protocol: TCP or UDP
	//    - Parse local/foreign address:port
	//    - Extract state from last column
	// 3. Return []Connection, nil
	//
	// Error handling:
	// - Return nil, fmt.Errorf("lsof failed: %w", err) on command failure
	// - Handle port parsing: strings.Split(addr, ":")[1] -> strconv.Atoi
	panic("implement GetConnections")
}

// Helper function to split ps output line into fields
//
// Ref: https://pkg.go.dev/strings
func parsePSLine(line string) (*Process, error) {
	// TODO: Implement parsePSLine
	//
	// Input line format (space-separated):
	// <pid> <ppid> <user> <%cpu> <%mem> <rss> <vsz> <state> <comm>
	//
	// Steps:
	// 1. Use strings.Fields(line) to split by whitespace
	// 2. Validate field count >= 9
	// 3. Parse each field:
	//    - pid, ppid: strconv.Atoi
	//    - cpu, mem: strconv.ParseFloat
	//    - rss, vsz: strconv.ParseUint (base 10, 0)
	//    - state: string (first char)
	//    - command: join remaining fields with space
	// 4. Return Process, nil
	//
	// Error handling:
	// - Return nil, fmt.Errorf("invalid line %q: %w", line, err) on parse failure
	panic("implement parsePSLine")
}

// Helper function to convert RSS bytes to human readable format
//
// Ref: https://pkg.go.dev/fmt
func FormatBytes(bytes uint64) string {
	// TODO: Implement FormatBytes
	//
	// Steps:
	// 1. Define units: B, KB, MB, GB, TB
	// 2. Loop: bytes > 1024 ? bytes /= 1024 : break
	// 3. Return fmt.Sprintf("%.1f%s", float64(bytes), unit)
	//
	// Examples:
	// - 1024 -> "1.0 KB"
	// - 1048576 -> "1.0 MB"
	// - 1073741824 -> "1.0 GB"
	panic("implement FormatBytes")
}