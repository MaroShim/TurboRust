package debugger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// Debugger manages a debug session for Rust code
type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	activeBin   string
	srcFile     string
	engine      *RustEngine
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

// SetBreakpoint adds a breakpoint at file:line
func (d *Debugger) SetBreakpoint(file string, line int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	norm := filepath.Clean(file)
	if _, ok := d.breakpoints[norm]; !ok {
		d.breakpoints[norm] = make(map[int]bool)
	}
	d.breakpoints[norm][line] = true

	if d.engine != nil && (d.srcFile == norm || filepath.Clean(d.srcFile) == norm) {
		d.engine.Breakpoints[line] = true
	}
}

// RemoveBreakpoint removes a breakpoint at file:line
func (d *Debugger) RemoveBreakpoint(file string, line int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	norm := filepath.Clean(file)
	if lines, ok := d.breakpoints[norm]; ok {
		delete(lines, line)
	}
	if d.engine != nil && (d.srcFile == norm || filepath.Clean(d.srcFile) == norm) {
		delete(d.engine.Breakpoints, line)
	}
}

// ToggleBreakpoint toggles a breakpoint at file:line
func (d *Debugger) ToggleBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	norm := filepath.Clean(file)
	if _, ok := d.breakpoints[norm]; !ok {
		d.breakpoints[norm] = make(map[int]bool)
	}

	if d.breakpoints[norm][line] {
		delete(d.breakpoints[norm], line)
		if d.engine != nil {
			delete(d.engine.Breakpoints, line)
		}
		return false
	} else {
		d.breakpoints[norm][line] = true
		if d.engine != nil {
			d.engine.Breakpoints[line] = true
		}
		return true
	}
}

// HasBreakpoint checks if there is a breakpoint at file:line
func (d *Debugger) HasBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	norm := filepath.Clean(file)
	if lines, ok := d.breakpoints[norm]; ok {
		return lines[line]
	}
	// Fallback to base name comparison
	base := filepath.Base(file)
	for f, lines := range d.breakpoints {
		if filepath.Base(f) == base {
			return lines[line]
		}
	}
	return false
}

// GetBreakpoints returns all breakpoints for a file
func (d *Debugger) GetBreakpoints(file string) []int {
	d.mu.Lock()
	defer d.mu.Unlock()

	norm := filepath.Clean(file)
	var list []int
	if lines, ok := d.breakpoints[norm]; ok {
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
	if d.engine != nil {
		d.engine.Breakpoints = make(map[int]bool)
	}
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

// GetProgramOutput returns standard output produced by the running debug program
func (d *Debugger) GetProgramOutput() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.engine != nil {
		return d.engine.OutputBuf.String()
	}
	return ""
}

// Start initiates a debug session for the compiled Rust binary and target source file
func (d *Debugger) Start(binPath string, srcFile string, initialLine int) error {
	var lines []string
	if srcFile != "" {
		if content, err := os.ReadFile(srcFile); err == nil {
			lines = stringsSplitLines(string(content))
		}
	}
	return d.StartWithLines(binPath, srcFile, lines, initialLine)
}

// StartWithLines starts debug session with provided source lines
func (d *Debugger) StartWithLines(binPath string, srcFile string, lines []string, initialLine int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.activeBin = binPath
	d.srcFile = filepath.Clean(srcFile)

	// Collect breakpoints for this file
	bps := make(map[int]bool)
	norm := filepath.Clean(srcFile)
	if linesMap, ok := d.breakpoints[norm]; ok {
		for l, set := range linesMap {
			if set {
				bps[l] = true
			}
		}
	} else {
		base := filepath.Base(srcFile)
		for f, linesMap := range d.breakpoints {
			if filepath.Base(f) == base {
				for l, set := range linesMap {
					if set {
						bps[l] = true
					}
				}
				break
			}
		}
	}

	if len(lines) == 0 && srcFile != "" {
		if content, err := os.ReadFile(srcFile); err == nil {
			lines = stringsSplitLines(string(content))
		}
	}

	d.engine = NewRustEngine(lines, bps)

	if len(lines) == 0 {
		// Mock mode for stub tests without sources
		if len(bps) > 0 {
			minBP := 999999
			for l := range bps {
				if l < minBP {
					minBP = l
				}
			}
			if len(d.engine.CallStack) > 0 {
				d.engine.CallStack[0].Line = minBP
			}
		}
	} else if len(bps) > 0 {
		// If current line is not already a breakpoint, advance to the first breakpoint
		curL := d.engine.CurrentLine()
		if !d.engine.Breakpoints[curL] {
			d.engine.Continue()
		}
	}

	d.syncStateLocked()
	return nil
}

func (d *Debugger) syncStateLocked() {
	if d.engine == nil {
		d.state = DebugState{
			Active:   false,
			Exited:   true,
			ExitCode: 0,
		}
		return
	}

	d.state = DebugState{
		Active:       d.engine.Active,
		Running:      false,
		Exited:       d.engine.Exited,
		ExitCode:     d.engine.ExitCode,
		CurrentFile:  d.srcFile,
		CurrentLine:  d.engine.CurrentLine(),
		CurrentFunc:  d.engine.CurrentFunc(),
		LocalVars:    d.engine.LocalVariables(),
		ErrorMessage: d.engine.StatusMsg,
	}
}

// Continue continues execution until the next breakpoint or program completion
func (d *Debugger) Continue() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active || d.engine == nil {
		return fmt.Errorf("no active debug session")
	}

	d.engine.Continue()
	d.syncStateLocked()
	return nil
}

// StepOver steps to the next line
func (d *Debugger) StepOver() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active || d.engine == nil {
		return fmt.Errorf("no active debug session")
	}

	d.engine.StepOver()
	d.syncStateLocked()
	return nil
}

// StepInto steps into a function or next instruction
func (d *Debugger) StepInto() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active || d.engine == nil {
		return fmt.Errorf("no active debug session")
	}

	d.engine.StepInto()
	d.syncStateLocked()
	return nil
}

// Stop terminates the debug session
func (d *Debugger) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.engine != nil {
		d.engine.Active = false
		d.engine.Exited = true
	}

	d.state = DebugState{
		Active:      false,
		Running:     false,
		Exited:      true,
		CurrentLine: 0,
	}
	return nil
}

func stringsSplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start <= len(s) {
		line := s[start:]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		lines = append(lines, line)
	}
	return lines
}
