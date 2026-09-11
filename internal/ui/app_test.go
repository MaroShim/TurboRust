package ui

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestAppMultiFileDebugging(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.rs")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.rs")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// 1. Set a breakpoint in math.rs:4
	if err := app.editor.LoadFile(calcMath); err != nil {
		t.Fatalf("failed to load math.rs: %v", err)
	}
	app.ToggleBreakpoint(4)

	// 2. Switch back to main.rs and set a breakpoint at line 14
	if err := app.editor.LoadFile(calcMain); err != nil {
		t.Fatalf("failed to load main.rs: %v", err)
	}
	app.ToggleBreakpoint(14)

	// Verify FileBreakpoints recorded both files
	if len(app.editor.FileBreakpoints[filepath.Clean(calcMath)]) == 0 {
		t.Errorf("expected math.rs breakpoints to be stored in FileBreakpoints")
	}

	// 3. Start debugging from main.rs
	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	// 4. Since factorial() is called first in main(), it stops at math.rs:4
	cleanMath := filepath.Clean(calcMath)
	if filepath.Clean(app.editor.FilePath) != cleanMath {
		t.Fatalf("expected editor to switch to math.rs, but is at %s", app.editor.FilePath)
	}

	// 5. Continue execution: should hit main.rs:14
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue failed: %v", err)
	}

	cleanMain := filepath.Clean(calcMain)
	if filepath.Clean(app.editor.FilePath) != cleanMain {
		t.Fatalf("expected editor to switch back to main.rs, but is at %s", app.editor.FilePath)
	}
}

func TestAppF7StepIntoMathRs(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.rs")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.rs")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// Set a single breakpoint at line 12 of main.rs (let fact = math::factorial(n);)
	app.ToggleBreakpoint(12)

	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	if app.editor.CurrentIP != 12 {
		t.Fatalf("expected CurrentIP 12, got %d", app.editor.CurrentIP)
	}

	// Press F7 (DebugStepInto) on line 13
	err = app.DebugStepInto()
	if err != nil {
		t.Fatalf("DebugStepInto failed: %v", err)
	}

	cleanMath := filepath.Clean(calcMath)
	if filepath.Clean(app.editor.FilePath) != cleanMath {
		t.Fatalf("expected editor to switch to math.rs, but is at %s", app.editor.FilePath)
	}
}

func TestAppTraceThroughFactorial(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.rs")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// Breakpoint at line 12 of main.rs (factorial call)
	app.ToggleBreakpoint(12)

	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	// Step into math::factorial
	t.Logf("Initial: IP=%d, File=%s", app.editor.CurrentIP, app.editor.FilePath)
	for step := 1; step <= 20; step++ {
		err := app.DebugStepInto()
		st := app.debugger.GetState()
		t.Logf("Step %d: err=%v, Active=%v, Exited=%v, File=%s, Line=%d, Func=%s, EditorFile=%s, EditorIP=%d",
			step, err, st.Active, st.Exited, st.CurrentFile, st.CurrentLine, st.CurrentFunc, filepath.Base(app.editor.FilePath), app.editor.CurrentIP)
		if !st.Active || st.Exited || err != nil {
			break
		}
	}
}

func TestAppMultipleFunctionBreakpoints(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.rs")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.rs")
	calcStats := filepath.Join(filepath.Dir(calcMain), "stats.rs")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// 1. Set breakpoints in math.rs
	if err := app.editor.LoadFile(calcMath); err != nil {
		t.Fatalf("failed to load math.rs: %v", err)
	}
	app.ToggleBreakpoint(4)  // factorial
	app.ToggleBreakpoint(16) // gcd
	app.ToggleBreakpoint(26) // is_prime
	app.ToggleBreakpoint(47) // power

	// 2. Set breakpoints in stats.rs
	if err := app.editor.LoadFile(calcStats); err != nil {
		t.Fatalf("failed to load stats.rs: %v", err)
	}
	app.ToggleBreakpoint(4)  // average
	app.ToggleBreakpoint(13) // variance
	app.ToggleBreakpoint(26) // std_dev
	app.ToggleBreakpoint(31) // min_max

	// 3. Switch back to main.rs and start debugging
	if err := app.editor.LoadFile(calcMain); err != nil {
		t.Fatalf("failed to load main.rs: %v", err)
	}

	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	// 4. Continue through all function breakpoints and record stops
	var hitStops []string
	for step := 0; step < 15; step++ {
		st := app.debugger.GetState()
		if !st.Active || st.Exited {
			break
		}
		cur := fmt.Sprintf("%s:%d", filepath.Base(app.editor.FilePath), app.editor.CurrentIP)
		hitStops = append(hitStops, cur)
		t.Logf("Hit stop #%d: %s (func=%s)", len(hitStops), cur, st.CurrentFunc)

		// Test F7 (Step Into) at the breakpoint
		if err := app.DebugStepInto(); err != nil {
			t.Fatalf("DebugStepInto failed at %s: %v", cur, err)
		}

		if err := app.DebugContinue(); err != nil {
			break
		}
	}

	t.Logf("Total stops hit: %d: %v", len(hitStops), hitStops)
	if len(hitStops) == 0 {
		t.Fatalf("expected multiple stops to be hit, got 0")
	}
}


