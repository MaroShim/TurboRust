package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CompileError holds parsed compiler diagnostic information
type CompileError struct {
	File    string
	Line    int
	Column  int
	Level   string // "error" or "warning"
	Message string
}

// BuildResult holds outcome of rustc / cargo build
type BuildResult struct {
	Success       bool
	LinesCompiled int
	Duration      time.Duration
	Errors        []CompileError
	ErrorCount    int
	WarningCount  int
	BinaryPath    string
	RawOutput     string
}

// RunResult holds outcome of program execution
type RunResult struct {
	Output    string
	ExitCode  int
	Duration  time.Duration
	Completed bool
}

// Regex for rustc error format
// 1. Short format: file.rs:line:col: error[...]: message
var shortErrRegex = regexp.MustCompile(`(?m)^([^:\n\r]+):(\d+):(\d+):\s*(error(?:\[\w+\])?|warning):\s*(.+)$`)

// 2. Standard rustc format:
// error[...]: message
//   --> file.rs:line:col
var stdLocRegex = regexp.MustCompile(`(?m)^\s*-->\s*([^:\n\r]+):(\d+):(\d+)`)
var stdMsgRegex = regexp.MustCompile(`(?m)^(error(?:\[\w+\])?|warning):\s*(.+)$`)

// CountLines counts total lines of Rust code in the specified target
func CountLines(targetPath string) int {
	total := 0
	info, err := os.Stat(targetPath)
	if err != nil {
		return 0
	}

	if !info.IsDir() {
		content, err := os.ReadFile(targetPath)
		if err == nil {
			total += bytes.Count(content, []byte("\n")) + 1
		}
		return total
	}

	_ = filepath.Walk(targetPath, func(path string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && strings.HasSuffix(path, ".rs") {
			content, err := os.ReadFile(path)
			if err == nil {
				total += bytes.Count(content, []byte("\n")) + 1
			}
		}
		return nil
	})
	return total
}

// ParseErrors extracts CompileError list from compiler output
func ParseErrors(output string) []CompileError {
	var errs []CompileError

	// First attempt short format matches
	shortMatches := shortErrRegex.FindAllStringSubmatch(output, -1)
	if len(shortMatches) > 0 {
		for _, m := range shortMatches {
			file := strings.TrimSpace(m[1])
			line, _ := strconv.Atoi(m[2])
			col, _ := strconv.Atoi(m[3])
			level := "error"
			if strings.HasPrefix(strings.ToLower(m[4]), "warning") {
				level = "warning"
			}
			msg := strings.TrimSpace(m[5])
			errs = append(errs, CompileError{
				File:    file,
				Line:    line,
				Column:  col,
				Level:   level,
				Message: msg,
			})
		}
		return errs
	}

	// Fallback to standard rustc multiline format
	lines := strings.Split(output, "\n")
	var lastLevel, lastMsg string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")

		msgMatch := stdMsgRegex.FindStringSubmatch(line)
		if len(msgMatch) > 0 {
			lvl := "error"
			if strings.HasPrefix(strings.ToLower(msgMatch[1]), "warning") {
				lvl = "warning"
			}
			lastLevel = lvl
			lastMsg = strings.TrimSpace(msgMatch[2])
			continue
		}

		locMatch := stdLocRegex.FindStringSubmatch(line)
		if len(locMatch) > 0 {
			file := strings.TrimSpace(locMatch[1])
			lNum, _ := strconv.Atoi(locMatch[2])
			colNum, _ := strconv.Atoi(locMatch[3])

			displayMsg := lastMsg
			if displayMsg == "" {
				displayMsg = "compiler diagnostic"
			}

			errs = append(errs, CompileError{
				File:    file,
				Line:    lNum,
				Column:  colNum,
				Level:   lastLevel,
				Message: displayMsg,
			})
			lastLevel = ""
			lastMsg = ""
		}
	}

	return errs
}

// FindCargoRoot checks if the target path or its ancestors contains Cargo.toml
func FindCargoRoot(targetPath string) (string, bool) {
	dir := targetPath
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		dir = filepath.Dir(dir)
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", false
	}

	for {
		cargoToml := filepath.Join(dir, "Cargo.toml")
		if _, err := os.Stat(cargoToml); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// Build compiles the target Rust file or Cargo package
func Build(targetPath string) *BuildResult {
	start := time.Now()
	res := &BuildResult{}

	lines := CountLines(targetPath)
	res.LinesCompiled = lines

	cargoRoot, hasCargo := FindCargoRoot(targetPath)
	_ = cargoRoot
	_ = hasCargo

	// For fast single file execution, prefer rustc with short error format
	dir := filepath.Dir(targetPath)
	if dir == "" {
		dir = "."
	}

	ext := ""
	if os.PathSeparator == '\\' {
		ext = ".exe"
	}
	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("turborust_bin_%d%s", time.Now().UnixNano(), ext))
	res.BinaryPath = tmpBin

	// Use rustc for single .rs file or if targetPath is a .rs file
	var cmd *exec.Cmd
	if strings.HasSuffix(targetPath, ".rs") {
		cmd = exec.Command("rustc", "--error-format=short", "-g", "-o", tmpBin, filepath.Base(targetPath))
		cmd.Dir = dir
	} else if hasCargo {
		cmd = exec.Command("cargo", "build", "--message-format=short")
		cmd.Dir = cargoRoot
	} else {
		cmd = exec.Command("rustc", "--error-format=short", "-g", "-o", tmpBin, targetPath)
		cmd.Dir = dir
	}

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.RawOutput = outBuf.String()

	if err != nil {
		res.Success = false
		res.Errors = ParseErrors(res.RawOutput)
		for _, e := range res.Errors {
			if e.Level == "error" {
				res.ErrorCount++
			} else {
				res.WarningCount++
			}
		}
		if res.ErrorCount == 0 && len(res.Errors) > 0 {
			res.ErrorCount = len(res.Errors)
		} else if res.ErrorCount == 0 {
			res.ErrorCount = 1
			res.Errors = append(res.Errors, CompileError{
				File:    filepath.Base(targetPath),
				Line:    1,
				Column:  1,
				Level:   "error",
				Message: strings.TrimSpace(res.RawOutput),
			})
		}
	} else {
		res.Success = true
		// Parse warnings if any
		warnings := ParseErrors(res.RawOutput)
		for _, w := range warnings {
			if w.Level == "warning" {
				res.WarningCount++
			}
		}
	}

	return res
}

// Run executes the compiled binary and returns output
func Run(binPath string, args ...string) *RunResult {
	start := time.Now()
	res := &RunResult{}

	cmd := exec.Command(binPath, args...)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.Output = outBuf.String()
	res.Completed = true

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = 1
		}
	} else {
		res.ExitCode = 0
	}

	return res
}
