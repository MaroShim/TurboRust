package debugger

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Regex patterns for parsing LLDB and GDB outputs
var (
	// LLDB frame: frame #0: 0x... [binary`]func at file.rs:18:19
	lldbFrameRe = regexp.MustCompile(`(?m)frame\s+#\d+:\s+0x[0-9a-fA-F]+\s+(?:[^\s\x60]+\x60)?([^\s(]+)?(?:\s*\([^)]*\))?\s+(?:at|at line)\s+([^:\s]+):(\d+)(?::(\d+))?`)
	// LLDB exit: Process 1234 exited with status = 0 (0x00000000)
	lldbExitRe = regexp.MustCompile(`(?m)Process\s+\d+\s+exited\s+with\s+status\s+=\s+(-?\d+)`)
	// LLDB variable: (type) name = val
	lldbVarRe = regexp.MustCompile(`^\(([^)]+)\)\s*([a-zA-Z_]\w*)\s*=\s*(.*)$`)

	// GDB frame: #0  main () at file.rs:18 or Breakpoint 1, main () at file.rs:18
	gdbFrameRe = regexp.MustCompile(`(?m)(?:#0\s+([^\s(]+)|Breakpoint\s+\d+,\s*([^\s(]+)).*at\s+([^:\s]+):(\d+)`)
	// GDB exit: [Inferior 1 (process 1234) exited normally] or exited with code 01
	gdbExitRe = regexp.MustCompile(`(?m)(?:exited normally|exited with code\s+(\d+))`)
	// GDB variable: name = val
	gdbVarRe = regexp.MustCompile(`^([a-zA-Z_]\w*)\s*=\s*(.*)$`)
)

// ExternalSession manages an interactive LLDB or GDB child process
type ExternalSession struct {
	mu           sync.Mutex
	debuggerPath string
	debuggerType string // "lldb" or "gdb"
	binPath      string
	srcFile      string
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	outChan      chan string
	doneChan     chan struct{}
	state        DebugState
	outputBuf    strings.Builder
}

// NewExternalSession launches lldb/gdb and sets initial breakpoints
func NewExternalSession(debuggerPath, debuggerType, binPath, srcFile string, breakpoints map[int]bool) (*ExternalSession, error) {
	var args []string
	if debuggerType == "gdb" {
		args = []string{"-q", binPath}
	} else {
		// lldb or rust-lldb
		args = []string{"--", binPath}
	}

	cmd := exec.Command(debuggerPath, args...)
	dir := filepath.Dir(binPath)
	if dir != "" {
		cmd.Dir = dir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe failed: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("stdout pipe failed: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("failed to start %s: %w", debuggerPath, err)
	}

	sess := &ExternalSession{
		debuggerPath: debuggerPath,
		debuggerType: debuggerType,
		binPath:      binPath,
		srcFile:      srcFile,
		cmd:          cmd,
		stdin:        stdin,
		outChan:      make(chan string, 500),
		doneChan:     make(chan struct{}),
		state: DebugState{
			Active:      true,
			Running:     false,
			Exited:      false,
			CurrentFile: srcFile,
			CurrentLine: 1,
		},
	}

	// Reader goroutine for stdout/stderr
	go func() {
		defer close(sess.doneChan)
		defer close(sess.outChan)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case sess.outChan <- line:
			case <-sess.doneChan:
				return
			}
		}
	}()

	// Configure debugger options and set breakpoints
	if debuggerType == "gdb" {
		sess.sendCmd("set pagination off")
		sess.sendCmd("set confirm off")
		for line := range breakpoints {
			sess.sendCmd(fmt.Sprintf("break %s:%d", filepath.Base(srcFile), line))
		}
		sess.sendCmd("run")
	} else {
		// LLDB / rust-lldb
		sess.sendCmd("settings set auto-confirm true")
		for line := range breakpoints {
			sess.sendCmd(fmt.Sprintf("breakpoint set -f %s -l %d", filepath.Base(srcFile), line))
		}
		sess.sendCmd("run")
	}

	// Wait for process to hit first breakpoint or start
	lines, err := sess.waitForStop(4 * time.Second)
	if err != nil {
		_ = sess.Stop()
		return nil, err
	}
	if !sess.state.Active || sess.state.Exited {
		_ = sess.Stop()
		return nil, fmt.Errorf("debugger process exited or failed to start")
	}

	sess.parseOutput(lines)
	sess.queryVariables()

	return sess, nil
}

func (s *ExternalSession) sendCmd(cmd string) {
	if s.stdin != nil {
		_, _ = io.WriteString(s.stdin, cmd+"\n")
	}
}

// waitForStop collects lines until a stop, breakpoint hit, or process exit event occurs
func (s *ExternalSession) waitForStop(timeout time.Duration) ([]string, error) {
	var collected []string
	deadline := time.After(timeout)

	stopKeywords := []string{
		"stopped", "stop reason", "Breakpoint", "breakpoint",
		"exited with status", "exited normally", "exited with code",
		"frame #", "* thread #",
	}

	for {
		select {
		case line, ok := <-s.outChan:
			if !ok {
				s.state.Active = false
				s.state.Exited = true
				return collected, fmt.Errorf("debugger process exited")
			}
			collected = append(collected, line)

			// Record user program output (filter out debugger prompt and command echoes)
			s.recordProgramOutput(line)

			// Check for exit
			if s.checkExit(line) {
				time.Sleep(30 * time.Millisecond)
				s.drainAvailableLines(&collected)
				return collected, nil
			}

			// Check for stop keywords
			for _, kw := range stopKeywords {
				if strings.Contains(line, kw) {
					// Allow trailing lines to arrive
					time.Sleep(50 * time.Millisecond)
					s.drainAvailableLines(&collected)
					return collected, nil
				}
			}

		case <-deadline:
			return collected, nil
		}
	}
}

func (s *ExternalSession) drainAvailableLines(collected *[]string) {
	for {
		select {
		case line, ok := <-s.outChan:
			if ok {
				*collected = append(*collected, line)
				s.recordProgramOutput(line)
				s.checkExit(line)
			} else {
				return
			}
		default:
			return
		}
	}
}

func (s *ExternalSession) recordProgramOutput(line string) {
	trimmed := strings.TrimSpace(line)
	// Ignore debugger control and status lines
	if strings.HasPrefix(trimmed, "(lldb)") ||
		strings.HasPrefix(trimmed, "(gdb)") ||
		strings.HasPrefix(trimmed, "Process ") ||
		strings.HasPrefix(trimmed, "Breakpoint ") ||
		strings.HasPrefix(trimmed, "Current executable") ||
		strings.HasPrefix(trimmed, "Target ") ||
		strings.HasPrefix(trimmed, "* thread #") ||
		strings.HasPrefix(trimmed, "frame #") ||
		strings.HasPrefix(trimmed, "->") ||
		strings.HasPrefix(trimmed, "warning:") ||
		strings.HasPrefix(trimmed, "Executing commands") {
		return
	}
	// Also ignore numbered code snippets (e.g. "   18        let fib = ...")
	if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' && strings.Contains(line, "\t") {
		return
	}
	if trimmed != "" {
		s.outputBuf.WriteString(line + "\n")
	}
}

func (s *ExternalSession) checkExit(line string) bool {
	if s.debuggerType == "gdb" {
		if m := gdbExitRe.FindStringSubmatch(line); len(m) > 0 {
			s.state.Active = false
			s.state.Exited = true
			if len(m) > 1 && m[1] != "" {
				code, _ := strconv.Atoi(m[1])
				s.state.ExitCode = code
			}
			return true
		}
	} else {
		if m := lldbExitRe.FindStringSubmatch(line); len(m) > 0 {
			s.state.Active = false
			s.state.Exited = true
			code, _ := strconv.Atoi(m[1])
			s.state.ExitCode = code
			return true
		}
	}
	return false
}

func (s *ExternalSession) parseOutput(lines []string) {
	for _, line := range lines {
		if s.debuggerType == "gdb" {
			if m := gdbFrameRe.FindStringSubmatch(line); len(m) > 0 {
				funcName := m[1]
				if funcName == "" && len(m) > 2 {
					funcName = m[2]
				}
				fileName := m[3]
				lineNo, _ := strconv.Atoi(m[4])
				s.state.CurrentFunc = funcName
				s.state.CurrentLine = lineNo
				if fileName != "" {
					s.state.CurrentFile = fileName
				}
				s.state.Running = false
				s.state.Active = true
				return
			}
		} else {
			if m := lldbFrameRe.FindStringSubmatch(line); len(m) > 0 {
				funcName := m[1]
				fileName := m[2]
				lineNo, _ := strconv.Atoi(m[3])
				s.state.CurrentFunc = funcName
				s.state.CurrentLine = lineNo
				if fileName != "" {
					s.state.CurrentFile = fileName
				}
				s.state.Running = false
				s.state.Active = true
				return
			}
		}
	}
}

// queryVariables queries locals from the current frame and parses them
func (s *ExternalSession) queryVariables() {
	if !s.state.Active || s.state.Exited {
		return
	}

	if s.debuggerType == "gdb" {
		s.sendCmd("info locals")
	} else {
		s.sendCmd("frame variable")
	}

	time.Sleep(100 * time.Millisecond)

	var varLines []string
	timeout := time.After(300 * time.Millisecond)
drainLoop:
	for {
		select {
		case line, ok := <-s.outChan:
			if !ok {
				break drainLoop
			}
			varLines = append(varLines, line)
		case <-timeout:
			break drainLoop
		}
	}

	var vars []Variable
	for _, l := range varLines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "(lldb)") || strings.HasPrefix(trimmed, "(gdb)") || trimmed == "" {
			continue
		}

		if s.debuggerType == "gdb" {
			if m := gdbVarRe.FindStringSubmatch(trimmed); len(m) >= 3 {
				vars = append(vars, Variable{
					Name:  strings.TrimSpace(m[1]),
					Type:  "var",
					Value: strings.TrimSpace(m[2]),
				})
			}
		} else {
			if m := lldbVarRe.FindStringSubmatch(trimmed); len(m) >= 4 {
				vars = append(vars, Variable{
					Type:  strings.TrimSpace(m[1]),
					Name:  strings.TrimSpace(m[2]),
					Value: strings.TrimSpace(m[3]),
				})
			} else if parts := strings.SplitN(trimmed, "=", 2); len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				vars = append(vars, Variable{
					Name:  name,
					Type:  "var",
					Value: val,
				})
			}
		}
	}

	if len(vars) > 0 {
		s.state.LocalVars = vars
	}
}

// Continue resumes program execution until the next breakpoint or exit
func (s *ExternalSession) Continue() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.state.Active || s.state.Exited {
		return fmt.Errorf("debugger is not active")
	}

	if s.debuggerType == "gdb" {
		s.sendCmd("continue")
	} else {
		s.sendCmd("thread continue")
	}

	lines, err := s.waitForStop(3 * time.Second)
	if err != nil {
		return err
	}
	s.parseOutput(lines)
	s.queryVariables()
	return nil
}

// StepOver executes the current line and pauses at the next line
func (s *ExternalSession) StepOver() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.state.Active || s.state.Exited {
		return fmt.Errorf("debugger is not active")
	}

	if s.debuggerType == "gdb" {
		s.sendCmd("next")
	} else {
		s.sendCmd("thread step-over")
	}

	lines, err := s.waitForStop(2 * time.Second)
	if err != nil {
		return err
	}
	s.parseOutput(lines)
	s.queryVariables()
	return nil
}

// StepInto steps into the function call on the current line
func (s *ExternalSession) StepInto() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.state.Active || s.state.Exited {
		return fmt.Errorf("debugger is not active")
	}

	if s.debuggerType == "gdb" {
		s.sendCmd("step")
	} else {
		s.sendCmd("thread step-in")
	}

	lines, err := s.waitForStop(2 * time.Second)
	if err != nil {
		return err
	}
	s.parseOutput(lines)
	s.queryVariables()
	return nil
}

// Stop terminates the debugger process
func (s *ExternalSession) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin != nil {
		if s.debuggerType == "gdb" {
			s.sendCmd("kill")
			s.sendCmd("quit")
		} else {
			s.sendCmd("process kill")
			s.sendCmd("quit")
		}
		_ = s.stdin.Close()
		s.stdin = nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
		s.cmd = nil
	}

	s.state.Active = false
	s.state.Running = false
	s.state.Exited = true

	return nil
}

func (s *ExternalSession) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state.Active && !s.state.Exited
}

func (s *ExternalSession) GetState() DebugState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *ExternalSession) GetOutput() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.outputBuf.String()
}
