package debugger

import (
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
