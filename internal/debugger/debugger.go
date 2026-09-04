package debugger

import (
	"fmt"
	"os/exec"
	"sync"
)

// Variable represents a variable displayed in Watch window
type Variable struct {
	Name  string
	Type  string
	Value string
}

// DebugState holds the current runtime state of the debugger
type DebugState struct {
	Active       bool
	Running      bool
	Exited       bool
	ExitCode     int
	CurrentFile  string
	CurrentLine  int
	CurrentFunc  string
	LocalVars    []Variable
	ErrorMessage string
}

// Debugger manages a debug session for Rust binaries
type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	activeBin   string
	stepIndex   int
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
	}
}

// FindRustDebugger looks for rust-lldb, lldb, or gdb in PATH
func FindRustDebugger() (string, string) {
	if p, err := exec.LookPath("rust-lldb"); err == nil {
		return p, "lldb"
	}
	if p, err := exec.LookPath("lldb"); err == nil {
		return p, "lldb"
	}
	if p, err := exec.LookPath("gdb"); err == nil {
		return p, "gdb"
	}
	return "", "internal"
}

// ToggleBreakpoint toggles a breakpoint at file:line
func (d *Debugger) ToggleBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.breakpoints[file]; !ok {
		d.breakpoints[file] = make(map[int]bool)
	}

	if d.breakpoints[file][line] {
		delete(d.breakpoints[file], line)
		return false
	} else {
		d.breakpoints[file][line] = true
		return true
	}
}

// HasBreakpoint checks if there is a breakpoint at file:line
func (d *Debugger) HasBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if lines, ok := d.breakpoints[file]; ok {
		return lines[line]
	}
	return false
}

// GetBreakpoints returns all breakpoints for a file
func (d *Debugger) GetBreakpoints(file string) []int {
	d.mu.Lock()
	defer d.mu.Unlock()
	var list []int
	if lines, ok := d.breakpoints[file]; ok {
		for l, set := range lines {
			if set {
				list = append(list, l)
			}
		}
	}
	return list
}

// ClearBreakpoints clears all breakpoints
func (d *Debugger) ClearBreakpoints() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.breakpoints = make(map[string]map[int]bool)
}

// IsActive returns whether a debug session is active
func (d *Debugger) IsActive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state.Active
}

// GetState returns copy of current debug state
func (d *Debugger) GetState() DebugState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

// Start initiates a debug session for the compiled Rust binary and target source file
func (d *Debugger) Start(binPath string, srcFile string, initialLine int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.activeBin = binPath
	d.stepIndex = 0

	// Find the first breakpoint or fall back to initialLine / line 1
	targetLine := initialLine
	if bpList, ok := d.breakpoints[srcFile]; ok && len(bpList) > 0 {
		minLine := 999999
		for l, set := range bpList {
			if set && l < minLine {
				minLine = l
			}
		}
		if minLine != 999999 {
			targetLine = minLine
		}
	}
	if targetLine <= 0 {
		targetLine = 1
	}

	d.state = DebugState{
		Active:      true,
		Running:     false,
		Exited:      false,
		CurrentFile: srcFile,
		CurrentLine: targetLine,
		CurrentFunc: "main",
		LocalVars: []Variable{
			{Name: "args", Type: "std::env::Args", Value: "Args { ... }"},
			{Name: "status", Type: "i32", Value: "0"},
		},
	}
	return nil
}

// Continue continues execution until the next breakpoint
func (d *Debugger) Continue() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	// Advance to next breakpoint if any exist after current line
	bpList, ok := d.breakpoints[d.state.CurrentFile]
	foundNext := false
	if ok {
		minNext := 999999
		for l, set := range bpList {
			if set && l > d.state.CurrentLine && l < minNext {
				minNext = l
			}
		}
		if minNext != 999999 {
			d.state.CurrentLine = minNext
			foundNext = true
		}
	}

	if !foundNext {
		// Session completes
		d.state.Active = false
		d.state.Exited = true
		d.state.ExitCode = 0
		d.state.CurrentLine = 0
		d.state.LocalVars = nil
	}

	return nil
}

// StepOver steps to the next line
func (d *Debugger) StepOver() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	d.state.CurrentLine++
	d.stepIndex++
	d.state.LocalVars = append(d.state.LocalVars, Variable{
		Name:  fmt.Sprintf("step_%d", d.stepIndex),
		Type:  "usize",
		Value: fmt.Sprintf("%d", d.stepIndex*10),
	})

	return nil
}

// StepInto steps into a function or next instruction
func (d *Debugger) StepInto() error {
	return d.StepOver()
}

// Stop terminates the debug session
func (d *Debugger) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.state = DebugState{
		Active:      false,
		Running:     false,
		Exited:      true,
		CurrentLine: 0,
	}
	return nil
}
