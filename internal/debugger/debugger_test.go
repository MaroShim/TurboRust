package debugger

import (
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
