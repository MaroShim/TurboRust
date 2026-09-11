package debugger

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebuggerBreakpoints(t *testing.T) {
	dbg := NewDebugger()
	file := "main.rs"

	if dbg.HasBreakpoint(file, 10) {
		t.Errorf("breakpoint should not exist yet")
	}

	set := dbg.ToggleBreakpoint(file, 10)
	if !set || !dbg.HasBreakpoint(file, 10) {
		t.Errorf("expected breakpoint at line 10 to be set")
	}

	set = dbg.ToggleBreakpoint(file, 10)
	if set || dbg.HasBreakpoint(file, 10) {
		t.Errorf("expected breakpoint at line 10 to be removed")
	}
}

func TestDebuggerSession(t *testing.T) {
	dbg := NewDebugger()
	file := "main.rs"
	dbg.ToggleBreakpoint(file, 5)

	err := dbg.Start("target/main.exe", file, 1)
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	state := dbg.GetState()
	if !state.Active || state.CurrentLine != 5 {
		t.Errorf("expected active at line 5, got %+v", state)
	}

	_ = dbg.StepOver()
	state = dbg.GetState()
	if state.CurrentLine != 6 {
		t.Errorf("expected step to line 6, got %d", state.CurrentLine)
	}

	_ = dbg.Stop()
	if dbg.IsActive() {
		t.Errorf("debugger should be inactive after stop")
	}
}

func TestRustEngineFibonacci(t *testing.T) {
	fibSource := `fn fibonacci(n: u32) -> u64 {
    if n <= 1 {
        return n as u64;
    }
    let mut a: u64 = 0;
    let mut b: u64 = 1;
    for _ in 2..=n {
        let temp = a + b;
        a = b;
        b = temp;
    }
    b
}

fn main() {
    println!("--- Fibonacci Sequence Calculator ---");
    for i in 0..15 {
        let fib = fibonacci(i);
        println!("F({:2}) = {}", i, fib);
    }
    println!("Done!");
}`

	lines := strings.Split(fibSource, "\n")
	dbg := NewDebugger()
	file := "fibonacci.rs"

	// Breakpoint at line 18: let fib = fibonacci(i);
	dbg.SetBreakpoint(file, 18)

	err := dbg.StartWithLines("fib.exe", file, lines, 1)
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	st := dbg.GetState()
	if st.CurrentLine != 18 {
		t.Errorf("expected stopped at line 18, got %d", st.CurrentLine)
	}

	// Step Into fibonacci
	err = dbg.StepInto()
	if err != nil {
		t.Fatalf("step into failed: %v", err)
	}
	st = dbg.GetState()
	if st.CurrentFunc != "fibonacci" {
		t.Errorf("expected current func fibonacci, got %s", st.CurrentFunc)
	}
	if st.CurrentLine != 2 {
		t.Errorf("expected inside fibonacci at line 2, got %d", st.CurrentLine)
	}

	// Step Over inside fibonacci (if n <= 1)
	_ = dbg.StepOver()
	st = dbg.GetState()
	if st.CurrentLine != 3 {
		t.Errorf("expected line 3 (return), got %d", st.CurrentLine)
	}

	// Step Over return -> back to main, line 19
	_ = dbg.StepOver()
	st = dbg.GetState()
	if st.CurrentFunc != "main" || st.CurrentLine != 19 {
		t.Errorf("expected back in main at line 19, got %s:%d", st.CurrentFunc, st.CurrentLine)
	}

	// Continue to next breakpoint hit (next iteration of loop, i=1)
	_ = dbg.Continue()
	st = dbg.GetState()
	if st.CurrentLine != 18 {
		t.Errorf("expected loop iteration to hit line 18, got %d", st.CurrentLine)
	}

	// Check variable i = 1
	foundI := false
	for _, v := range st.LocalVars {
		if v.Name == "i" {
			foundI = true
			if v.Value != "1" {
				t.Errorf("expected i = 1, got %s", v.Value)
			}
		}
	}
	if !foundI {
		t.Errorf("variable i not found in local vars: %+v", st.LocalVars)
	}

	// Continue to i=2
	_ = dbg.Continue()
	st = dbg.GetState()
	if st.CurrentLine != 18 {
		t.Errorf("expected line 18, got %d", st.CurrentLine)
	}
	for _, v := range st.LocalVars {
		if v.Name == "i" && v.Value != "2" {
			t.Errorf("expected i = 2, got %s", v.Value)
		}
	}

	_ = dbg.Stop()
}

func TestRustEngineHello(t *testing.T) {
	helloSource := `fn main() {
    println!("Hello, Turbo Rust World!");
}`
	lines := strings.Split(helloSource, "\n")
	dbg := NewDebugger()
	file := "hello.rs"

	err := dbg.StartWithLines("hello.exe", file, lines, 1)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	st := dbg.GetState()
	if st.CurrentLine != 2 {
		t.Errorf("expected start at line 2, got %d", st.CurrentLine)
	}

	_ = dbg.StepOver()
	out := dbg.GetProgramOutput()
	if !strings.Contains(out, "Hello, Turbo Rust World!") {
		t.Errorf("expected hello output, got %q", out)
	}
}

func TestFindRustDebugger(t *testing.T) {
	path, dbgType := FindRustDebugger()
	t.Logf("Detected debugger: path=%s, type=%s", path, dbgType)
	if path != "" {
		if dbgType != "lldb" && dbgType != "gdb" {
			t.Errorf("unexpected debugger type %s for path %s", dbgType, path)
		}
	} else {
		if dbgType != "internal" {
			t.Errorf("expected internal fallback when no debugger found, got %s", dbgType)
		}
	}
}

func TestInternalDebuggerFallbackWhenNoBinary(t *testing.T) {
	dbg := NewDebugger()
	lines := []string{
		"fn main() {",
		"    let x = 42;",
		"}",
	}
	// Non-existent binary path should fall back to internal
	err := dbg.StartWithLines("/non/existent/binary_path_12345", "main.rs", lines, 1)
	if err != nil {
		t.Fatalf("StartWithLines failed: %v", err)
	}
	if dbg.BackendType() != "internal" {
		t.Errorf("expected internal fallback for non-existent binary, got %s", dbg.BackendType())
	}
	_ = dbg.Stop()
}

func TestNativeDebuggerExecution(t *testing.T) {
	dbgPath, dbgType := FindRustDebugger()
	if dbgType == "internal" {
		t.Skip("No native debugger (lldb/gdb) found on system, skipping native execution test")
	}

	// Verify rustc exists
	rustcPath, err := exec.LookPath("rustc")
	if err != nil {
		t.Skip("rustc not found, skipping native execution test")
	}

	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "test_dbg.rs")
	binFile := filepath.Join(tempDir, "test_dbg_bin")

	srcCode := `fn main() {
    let mut count = 0;
    count += 1;
    println!("count={}", count);
}
`
	if err := os.WriteFile(srcFile, []byte(srcCode), 0644); err != nil {
		t.Fatal(err)
	}

	// Compile with debug symbols
	cmd := exec.Command(rustcPath, "-g", "-o", binFile, srcFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("rustc compilation failed: %v, output: %s", err, string(out))
	}

	dbg := NewDebugger()
	// Breakpoint at line 3: count += 1;
	dbg.SetBreakpoint(srcFile, 3)

	lines := strings.Split(srcCode, "\n")
	if err := dbg.StartWithLines(binFile, srcFile, lines, 1); err != nil {
		t.Fatalf("StartWithLines failed: %v", err)
	}

	backend := dbg.BackendType()
	t.Logf("Active backend: %s (path: %s)", backend, dbgPath)
	if backend == "internal" {
		t.Errorf("expected native backend, got internal")
	}

	st := dbg.GetState()
	t.Logf("Initial stop state: file=%s line=%d func=%s", st.CurrentFile, st.CurrentLine, st.CurrentFunc)

	// Step over
	if err := dbg.StepOver(); err != nil {
		t.Fatalf("StepOver failed: %v", err)
	}
	st = dbg.GetState()
	t.Logf("State after StepOver: line=%d vars=%+v", st.CurrentLine, st.LocalVars)

	// Continue to exit
	_ = dbg.Continue()
	st = dbg.GetState()
	t.Logf("Final state: exited=%v exitCode=%d", st.Exited, st.ExitCode)

	_ = dbg.Stop()
}
