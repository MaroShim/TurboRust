package dialogs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"tr/internal/compiler"
)

func newSimScreen(t *testing.T) tcell.Screen {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	s.SetSize(80, 25)
	return s
}

func TestAboutDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewAboutDialog()

	if d.IsVisible() {
		t.Errorf("expected initially not visible")
	}

	d.Show()
	if !d.IsVisible() {
		t.Errorf("expected visible after Show()")
	}

	d.Draw(screen, 80, 25)

	d.Hide()
	if d.IsVisible() {
		t.Errorf("expected not visible after Hide()")
	}
}

func TestCompileDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewCompileDialog()

	if d.IsVisible() {
		t.Errorf("expected initially not visible")
	}

	res := &compiler.BuildResult{
		Success:       true,
		LinesCompiled: 150,
		Duration:      120 * time.Millisecond,
	}

	d.Show("main.go", 150, res)
	if !d.IsVisible() || d.FileName != "main.go" || d.Lines != 150 {
		t.Errorf("unexpected state after Show: %+v", d)
	}

	d.Draw(screen, 80, 25)

	dismissed := false
	d.OnDismiss = func() {
		dismissed = true
	}

	d.Hide()
	if d.IsVisible() || !dismissed {
		t.Errorf("expected dismissed callback triggered and visible false")
	}
}

func TestErrorListDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewErrorListDialog()

	errs := []compiler.CompileError{
		{File: "main.go", Line: 10, Column: 5, Level: "error", Message: "undefined: x"},
		{File: "main.go", Line: 20, Column: 1, Level: "warning", Message: "unused variable"},
	}

	jumped := false
	var jumpedErr compiler.CompileError
	d.Show(errs, func(err compiler.CompileError) {
		jumped = true
		jumpedErr = err
	})

	if !d.IsVisible() || len(d.Errors) != 2 || d.SelectedIndex != 0 {
		t.Fatalf("unexpected state after Show: %+v", d)
	}

	d.Draw(screen, 80, 25)

	d.MoveDown()
	if d.SelectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", d.SelectedIndex)
	}

	// MoveDown boundary
	d.MoveDown()
	if d.SelectedIndex != 1 {
		t.Errorf("expected selectedIndex to stay at 1, got %d", d.SelectedIndex)
	}

	d.MoveUp()
	if d.SelectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", d.SelectedIndex)
	}

	d.SelectCurrent()
	if !jumped || jumpedErr.Line != 10 {
		t.Errorf("expected jump to line 10, got %+v", jumpedErr)
	}
	if d.IsVisible() {
		t.Errorf("expected dialog to hide after SelectCurrent")
	}
}

func TestFindDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewFindDialog()

	var foundQuery string
	var foundCase bool
	d.ShowWithTitle("Find in Project", "hello", func(q string, cs bool) {
		foundQuery = q
		foundCase = cs
	})

	if !d.IsVisible() || d.Title != "Find in Project" || d.Query != "hello" {
		t.Errorf("unexpected FindDialog state: %+v", d)
	}

	d.InsertRune('!')
	if d.Query != "hello!" {
		t.Errorf("expected 'hello!', got %q", d.Query)
	}

	d.Backspace()
	if d.Query != "hello" {
		t.Errorf("expected 'hello', got %q", d.Query)
	}

	// NextField to CaseSensitive checkbox
	d.NextField()
	if d.FocusField != 1 {
		t.Errorf("expected focusField 1, got %d", d.FocusField)
	}
	d.InsertRune(' ') // Space toggles checkbox
	if !d.CaseSensitive {
		t.Errorf("expected CaseSensitive true after space")
	}

	d.PrevField()
	if d.FocusField != 0 {
		t.Errorf("expected focusField 0, got %d", d.FocusField)
	}

	d.Draw(screen, 80, 25)

	d.Confirm()
	if foundQuery != "hello" || !foundCase {
		t.Errorf("expected query 'hello' with case=true, got query=%q case=%v", foundQuery, foundCase)
	}
	if d.IsVisible() {
		t.Errorf("expected dialog to hide after Confirm")
	}
}

func TestGotoLineDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewGotoLineDialog()

	jumpedLine := -1
	d.Show(1, 100, func(line int) {
		jumpedLine = line
	})

	if !d.IsVisible() || d.TotalLines != 100 {
		t.Errorf("unexpected state: %+v", d)
	}

	d.InsertRune('a') // Non-digit should be ignored
	if d.LineText != "1" {
		t.Errorf("expected non-digit ignored, got %q", d.LineText)
	}

	d.InsertRune('5')
	if d.LineText != "15" {
		t.Errorf("expected '15', got %q", d.LineText)
	}

	d.Backspace()
	if d.LineText != "1" {
		t.Errorf("expected '1', got %q", d.LineText)
	}

	d.Draw(screen, 80, 25)

	d.Confirm()
	if jumpedLine != 1 {
		t.Errorf("expected jump to line 1, got %d", jumpedLine)
	}
	if d.IsVisible() {
		t.Errorf("expected dialog to hide after Confirm")
	}
}

func TestSaveFileDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewSaveFileDialog()

	savedPath := ""
	d.Show("test.go", func(path string) {
		savedPath = path
	})

	if !d.IsVisible() || d.FileName != "test.go" {
		t.Errorf("unexpected state: %+v", d)
	}

	d.InsertRune('2')
	if d.FileName != "test.go2" {
		t.Errorf("expected test.go2, got %q", d.FileName)
	}

	d.Backspace()
	if d.FileName != "test.go" {
		t.Errorf("expected test.go, got %q", d.FileName)
	}

	d.Draw(screen, 80, 25)

	d.Confirm()
	if savedPath != "test.go" {
		t.Errorf("expected savedPath 'test.go', got %q", savedPath)
	}
	if d.IsVisible() {
		t.Errorf("expected dialog to hide after Confirm")
	}
}

func TestOpenFileDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewOpenFileDialog()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "sample.go")
	_ = os.WriteFile(testFile, []byte("package main"), 0644)

	openedPath := ""
	d.Show(tempDir, func(path string) {
		openedPath = path
	})

	if !d.IsVisible() || len(d.Files) == 0 {
		t.Fatalf("expected files to be listed, got %+v", d.Files)
	}

	d.Draw(screen, 80, 25)

	// Move down and handle enter
	d.MoveDown()
	d.MoveUp()

	d.Hide()
	if d.IsVisible() {
		t.Errorf("expected not visible after Hide()")
	}
	_ = openedPath
}

func TestSearchResultsDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewSearchResultsDialog()

	matches := []compiler.SearchMatch{
		{File: "/tmp/foo.go", Line: 10, Column: 4, Snippet: "func foo() {"},
		{File: "/tmp/bar.go", Line: 20, Column: 2, Snippet: "func bar() {"},
	}

	var jumpedMatch compiler.SearchMatch
	d.Show(matches, "/tmp", func(m compiler.SearchMatch) {
		jumpedMatch = m
	})

	if !d.IsVisible() || len(d.Matches) != 2 {
		t.Fatalf("unexpected state: %+v", d)
	}

	d.Draw(screen, 80, 25)

	d.MoveDown()
	if d.SelectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", d.SelectedIndex)
	}

	d.SelectCurrent()
	if jumpedMatch.File != "/tmp/bar.go" || jumpedMatch.Line != 20 {
		t.Errorf("expected jump to bar.go:20, got %+v", jumpedMatch)
	}
	if d.IsVisible() {
		t.Errorf("expected dialog to hide after SelectCurrent")
	}
}

func TestConfirmSaveDialog(t *testing.T) {
	screen := newSimScreen(t)
	d := NewConfirmSaveDialog()

	if d.IsVisible() {
		t.Errorf("expected initially not visible")
	}

	var selectedChoice ConfirmChoice = -1
	d.Show("test.rs", func(choice ConfirmChoice) {
		selectedChoice = choice
	})

	if !d.IsVisible() || d.FileName != "test.rs" || d.SelectedIndex != 0 {
		t.Fatalf("unexpected state after Show: %+v", d)
	}

	d.Draw(screen, 80, 25)

	// Test navigation
	d.MoveRight()
	if d.SelectedIndex != 1 {
		t.Errorf("expected selected 1 after MoveRight, got %d", d.SelectedIndex)
	}
	d.MoveRight()
	if d.SelectedIndex != 2 {
		t.Errorf("expected selected 2 after second MoveRight, got %d", d.SelectedIndex)
	}
	d.MoveRight()
	if d.SelectedIndex != 0 {
		t.Errorf("expected wrapped around to 0, got %d", d.SelectedIndex)
	}
	d.MoveLeft()
	if d.SelectedIndex != 2 {
		t.Errorf("expected wrapped around to 2 with MoveLeft, got %d", d.SelectedIndex)
	}

	// Test Confirm
	d.Confirm()
	if d.IsVisible() {
		t.Errorf("expected dialog hidden after Confirm")
	}
	if selectedChoice != ConfirmCancel {
		t.Errorf("expected ConfirmCancel, got %v", selectedChoice)
	}

	// Test Choose Yes
	d.Show("main.rs", func(choice ConfirmChoice) {
		selectedChoice = choice
	})
	d.Choose(ConfirmYes)
	if d.IsVisible() || selectedChoice != ConfirmYes {
		t.Errorf("expected Choose Yes to work, got visible=%v choice=%v", d.IsVisible(), selectedChoice)
	}

	// Test Choose No
	d.Show("main.rs", func(choice ConfirmChoice) {
		selectedChoice = choice
	})
	d.Choose(ConfirmNo)
	if d.IsVisible() || selectedChoice != ConfirmNo {
		t.Errorf("expected Choose No to work, got visible=%v choice=%v", d.IsVisible(), selectedChoice)
	}
}
