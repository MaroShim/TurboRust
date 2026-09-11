package debugger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	extSession  *ExternalSession
	backendType string
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
		backendType: "internal",
	}
}

// BackendType returns the currently active debugger backend ("rust-lldb", "lldb", "gdb", or "internal")
func (d *Debugger) BackendType() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.backendType != "" {
		return d.backendType
	}
	return "internal"
}

func isDebuggerFunctional(cmdPath string) bool {
	cmd := exec.Command(cmdPath, "--version")
	return cmd.Run() == nil
}

// FindRustDebugger looks for a functional debugger matching OS conventions:
// - Linux: gdb / rust-gdb prioritized as the primary Linux debugging backend
// - macOS: lldb / rust-lldb prioritized as standard macOS debugging backend
func FindRustDebugger() (string, string) {
	var candidates []struct {
		name    string
		dbgType string
	}

	if runtime.GOOS == "darwin" {
		candidates = []struct {
			name    string
			dbgType string
		}{
			{"rust-lldb", "lldb"},
			{"lldb", "lldb"},
			{"rust-gdb", "gdb"},
			{"gdb", "gdb"},
		}
	} else {
		// Linux & Windows: prioritize GDB first
		candidates = []struct {
			name    string
			dbgType string
		}{
			{"rust-gdb", "gdb"},
			{"gdb", "gdb"},
			{"rust-lldb", "lldb"},
			{"lldb", "lldb"},
		}
	}

	for _, c := range candidates {
		if p, err := exec.LookPath(c.name); err == nil && isDebuggerFunctional(p) {
			return p, c.dbgType
		}
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
	if d.extSession != nil {
		if d.extSession.debuggerType == "gdb" {
			d.extSession.sendCmd(fmt.Sprintf("break %s:%d", filepath.Base(file), line))
		} else {
			d.extSession.sendCmd(fmt.Sprintf("breakpoint set -f %s -l %d", filepath.Base(file), line))
		}
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
	if d.extSession != nil {
		if d.extSession.debuggerType == "gdb" {
			d.extSession.sendCmd(fmt.Sprintf("clear %s:%d", filepath.Base(file), line))
		} else {
			d.extSession.sendCmd(fmt.Sprintf("breakpoint clear -f %s -l %d", filepath.Base(file), line))
		}
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
		if d.extSession != nil {
			if d.extSession.debuggerType == "gdb" {
				d.extSession.sendCmd(fmt.Sprintf("clear %s:%d", filepath.Base(file), line))
			} else {
				d.extSession.sendCmd(fmt.Sprintf("breakpoint clear -f %s -l %d", filepath.Base(file), line))
			}
		}
		return false
	} else {
		d.breakpoints[norm][line] = true
		if d.engine != nil {
			d.engine.Breakpoints[line] = true
		}
		if d.extSession != nil {
			if d.extSession.debuggerType == "gdb" {
				d.extSession.sendCmd(fmt.Sprintf("break %s:%d", filepath.Base(file), line))
			} else {
				d.extSession.sendCmd(fmt.Sprintf("breakpoint set -f %s -l %d", filepath.Base(file), line))
			}
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
	if d.extSession != nil {
		if d.extSession.debuggerType == "gdb" {
			d.extSession.sendCmd("delete")
		} else {
			d.extSession.sendCmd("breakpoint delete")
		}
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
	if d.extSession != nil {
		return d.extSession.GetOutput()
	}
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

	// Clean up any existing session
	if d.extSession != nil {
		_ = d.extSession.Stop()
		d.extSession = nil
	}
	if d.engine != nil {
		d.engine = nil
	}

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

	// 1. Try external debugger (rust-lldb, lldb, gdb) if binary exists
	if binPath != "" {
		if _, err := os.Stat(binPath); err == nil {
			if dbgPath, dbgType := FindRustDebugger(); dbgType != "internal" {
				ext, err := NewExternalSession(dbgPath, dbgType, binPath, srcFile, bps)
				if err == nil && ext != nil && ext.IsActive() {
					d.extSession = ext
					d.backendType = filepath.Base(dbgPath)
					d.syncStateLocked()
					return nil
				}
				if ext != nil {
					_ = ext.Stop()
				}
			}
		}
	}

	// 2. Fallback to internal RustEngine
	d.backendType = "internal"
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
	if d.extSession != nil {
		d.state = d.extSession.GetState()
		return
	}
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

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	if d.extSession != nil && d.extSession.IsActive() {
		err := d.extSession.Continue()
		d.syncStateLocked()
		return err
	}

	if d.engine == nil {
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

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	if d.extSession != nil && d.extSession.IsActive() {
		err := d.extSession.StepOver()
		d.syncStateLocked()
		return err
	}

	if d.engine == nil {
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

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	if d.extSession != nil && d.extSession.IsActive() {
		err := d.extSession.StepInto()
		d.syncStateLocked()
		return err
	}

	if d.engine == nil {
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

	if d.extSession != nil {
		_ = d.extSession.Stop()
		d.extSession = nil
	}

	if d.engine != nil {
		d.engine.Active = false
		d.engine.Exited = true
		d.engine = nil
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
