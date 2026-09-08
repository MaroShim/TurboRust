package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

// Regex for rustc error format (supports Windows drive letters C:\...)
// 1. Short format: file.rs:line:col: error[...]: message
var shortErrRegex = regexp.MustCompile(`(?m)^((?:[a-zA-Z]:)?[^:\n\r]+):(\d+):(\d+):\s*(error(?:\[\w+\])?|warning):\s*(.+)$`)

// 2. Standard rustc format:
// error[...]: message
//   --> file.rs:line:col
var stdLocRegex = regexp.MustCompile(`(?m)^\s*-->\s*((?:[a-zA-Z]:)?[^:\n\r]+):(\d+):(\d+)`)
var stdMsgRegex = regexp.MustCompile(`(?m)^(error(?:\[\w+\])?|warning):\s*(.+)$`)

// GetCargoPackageName reads the package name from Cargo.toml
func GetCargoPackageName(cargoRoot string) string {
	cargoToml := filepath.Join(cargoRoot, "Cargo.toml")
	data, err := os.ReadFile(cargoToml)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`(?m)^\s*name\s*=\s*["']([^"']+)["']`)
	m := re.FindStringSubmatch(string(data))
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// CountLines counts total lines of Rust code in the specified target, package, or Cargo project
func CountLines(targetPath string) int {
	total := 0
	info, err := os.Stat(targetPath)
	if err != nil {
		return 0
	}

	if info.IsDir() {
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

	absTarget, _ := filepath.Abs(targetPath)
	dir := filepath.Dir(absTarget)

	// If inside a Cargo project, count all .rs files in cargoRoot (excluding target/)
	if cargoRoot, hasCargo := FindCargoRoot(absTarget); hasCargo {
		_ = filepath.Walk(cargoRoot, func(path string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && strings.HasSuffix(path, ".rs") {
				if !strings.Contains(path, string(filepath.Separator)+"target"+string(filepath.Separator)) {
					content, err := os.ReadFile(path)
					if err == nil {
						total += bytes.Count(content, []byte("\n")) + 1
					}
				}
			}
			return nil
		})
		if total > 0 {
			return total
		}
	}

	// Standalone multi-file directory
	entries, err := os.ReadDir(dir)
	if err == nil {
		fileCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rs") {
				content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
				if err == nil {
					total += bytes.Count(content, []byte("\n")) + 1
					fileCount++
				}
			}
		}
		if fileCount > 0 {
			return total
		}
	}

	content, err := os.ReadFile(targetPath)
	if err == nil {
		total = bytes.Count(content, []byte("\n")) + 1
	}
	return total
}

// ParseErrors extracts CompileError list from compiler output, resolving relative paths with workDir
func ParseErrors(output string, workDirs ...string) []CompileError {
	var errs []CompileError
	workDir := ""
	if len(workDirs) > 0 {
		workDir = workDirs[0]
	}

	resolvePath := func(f string) string {
		if !filepath.IsAbs(f) && workDir != "" {
			return filepath.Clean(filepath.Join(workDir, f))
		}
		return f
	}

	// First attempt short format matches
	shortMatches := shortErrRegex.FindAllStringSubmatch(output, -1)
	if len(shortMatches) > 0 {
		for _, m := range shortMatches {
			file := resolvePath(strings.TrimSpace(m[1]))
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
			file := resolvePath(strings.TrimSpace(locMatch[1]))
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

	absTarget, _ := filepath.Abs(targetPath)
	dir := filepath.Dir(absTarget)
	if dir == "" {
		dir = "."
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	cargoRoot, hasCargo := FindCargoRoot(absTarget)
	var cmd *exec.Cmd
	workDir := dir

	if hasCargo {
		// Cargo project has top priority
		workDir = cargoRoot
		cmd = exec.Command("cargo", "build", "--message-format=short")
		cmd.Dir = cargoRoot

		pkgName := GetCargoPackageName(cargoRoot)
		if pkgName == "" {
			pkgName = filepath.Base(cargoRoot)
		}
		res.BinaryPath = filepath.Join(cargoRoot, "target", "debug", pkgName+ext)
	} else {
		// Single or multi-file standalone Rust project
		workDir = dir
		tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("turborust_bin_%d%s", time.Now().UnixNano(), ext))
		res.BinaryPath = tmpBin
		cmd = exec.Command("rustc", "--error-format=short", "-g", "-o", tmpBin, filepath.Base(absTarget))
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
		res.Errors = ParseErrors(res.RawOutput, workDir)
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
		warnings := ParseErrors(res.RawOutput, workDir)
		for _, w := range warnings {
			if w.Level == "warning" {
				res.WarningCount++
			}
		}
	}

	return res
}

// BuildDebug compiles target Rust file or Cargo package with debug symbols
func BuildDebug(targetPath string) *BuildResult {
	// Build() already creates debug binaries (-g for rustc, cargo build produces debug by default)
	return Build(targetPath)
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
