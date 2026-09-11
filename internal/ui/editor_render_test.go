package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newSimScreen(t *testing.T, w, h int) tcell.Screen {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	s.SetSize(w, h)
	return s
}

func TestEditorDraw_FocusedCursorAndGutter(t *testing.T) {
	screen := newSimScreen(t, 80, 25)

	ed := NewEditor("main.go", 1)
	ed.Lines = []string{
		"package main",
		"func main() {",
		"\tprintln(\"hello\")",
		"}",
	}
	ed.ShowLineNums = true
	ed.CursorY = 1 // line 2: "func main() {"
	ed.CursorX = 5 // on 'm'

	// Draw focused editor
	ed.Draw(screen, 0, 1, 80, 20, true)

	// 1. Check Gutter at row 2 (screenY = 1 + 1 = 2):
	// Must contain the cursor indicator '▸'
	r, _, _, _ := screen.GetContent(0, 2)
	if r != '▸' {
		t.Errorf("expected gutter marker '▸' on current line, got %q (rune %d)", r, r)
	}

	// Other lines (e.g. line 0, screenY = 1) must NOT contain '▸'
	r0, _, _, _ := screen.GetContent(0, 1)
	if r0 == '▸' {
		t.Errorf("expected non-cursor line gutter to NOT have '▸', got %q", r0)
	}

	// 2. Check Hardware Cursor:
	// Screen coordinates for cursor:
	// lineNumWidth: digits + 2 = 3 + 2 = 5
	// Cursor is at col 5, so screenX = 5 + 5 = 10, screenY = 2
	cellRune, _, style, _ := screen.GetContent(10, 2)
	if cellRune != 'm' {
		t.Errorf("expected cursor cell to contain 'm', got %q", cellRune)
	}
	_, bg, _ := style.Decompose()
	if bg != ColorEditorBg {
		t.Errorf("expected cursor cell to have normal background %v, got %v", ColorEditorBg, bg)
	}
}

func TestEditorDraw_UnfocusedState(t *testing.T) {
	screen := newSimScreen(t, 80, 25)

	ed := NewEditor("main.go", 1)
	ed.Lines = []string{
		"package main",
		"func main() {",
	}
	ed.ShowLineNums = true
	ed.CursorY = 1
	ed.CursorX = 5

	// Draw unfocused editor (e.g., when dialog or menu is active)
	ed.Draw(screen, 0, 1, 80, 20, false)

	// Gutter must NOT have '▸' when unfocused
	r, _, _, _ := screen.GetContent(0, 2)
	if r == '▸' {
		t.Errorf("expected no '▸' gutter marker when unfocused, got %q", r)
	}

	// Cursor cell has normal editor background
	cellRune, _, style, _ := screen.GetContent(10, 2)
	if cellRune != 'm' {
		t.Errorf("expected 'm', got %q", cellRune)
	}
	_, bg, _ := style.Decompose()
	if bg != ColorEditorBg {
		t.Errorf("expected ColorEditorBg when unfocused, got %v", bg)
	}
}

func TestEditorDraw_TabAndEolCursor(t *testing.T) {
	screen := newSimScreen(t, 80, 25)

	ed := NewEditor("main.go", 1)
	ed.Lines = []string{
		"\tprintln(\"hi\")",
		"",
	}
	ed.ShowLineNums = true
	ed.TabWidth = 4
	ed.CursorY = 1 // Empty line
	ed.CursorX = 0

	// Draw focused on empty line
	ed.Draw(screen, 0, 1, 80, 20, true)

	// On empty line, cursor cell at screenX = 5 (after 5-char gutter) must be space with ColorEditorBg
	r, _, style, _ := screen.GetContent(5, 2)
	if r != ' ' {
		t.Errorf("expected space at EOL cursor, got %q", r)
	}
	_, bg, _ := style.Decompose()
	if bg != ColorEditorBg {
		t.Errorf("expected ColorEditorBg on empty line cursor, got %v", bg)
	}
}
