package ui

import (
	"testing"
)

func TestEditorOperations(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{""}
	ed.CursorX = 0
	ed.CursorY = 0

	// 1. Insert characters
	for _, r := range "func foo() {" {
		ed.InsertRune(r)
	}
	if ed.Lines[0] != "func foo() {" {
		t.Errorf("expected 'func foo() {', got %q", ed.Lines[0])
	}
	if ed.CursorX != 12 {
		t.Errorf("expected cursorX 12, got %d", ed.CursorX)
	}

	// 2. Insert new line
	ed.InsertNewLine()
	if len(ed.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(ed.Lines))
	}
	if ed.CursorY != 1 {
		t.Errorf("expected cursorY 1, got %d", ed.CursorY)
	}

	// 3. Insert tab (4 spaces to reach tab stop 4)
	ed.InsertTab()
	if ed.CursorX != 4 || ed.Lines[1] != "    " {
		t.Errorf("expected 4 spaces, got %q (cursorX=%d)", ed.Lines[1], ed.CursorX)
	}

	// 4. Smart Backspace: deletes 4 spaces indentation at once
	ed.Backspace()
	if ed.CursorX != 0 || ed.Lines[1] != "" {
		t.Errorf("expected 0 spaces after smart backspace, got %q (cursorX=%d)", ed.Lines[1], ed.CursorX)
	}

	// Test ExpandTabs
	tabbed := "\tfmt.Println(\"ok\")"
	expanded := ExpandTabs(tabbed, 4)
	if expanded != "    fmt.Println(\"ok\")" {
		t.Errorf("expected expanded tab to 4 spaces, got %q", expanded)
	}

	// 5. Breakpoint toggling
	isBp := ed.ToggleBreakpoint(1)
	if !isBp || !ed.Breakpoints[1] {
		t.Errorf("expected breakpoint at line 1 to be set")
	}
	isBp = ed.ToggleBreakpoint(1)
	if isBp || ed.Breakpoints[1] {
		t.Errorf("expected breakpoint at line 1 to be removed")
	}

	// 6. GotoLine
	ed.GotoLine(1, 6)
	if ed.CursorY != 0 || ed.CursorX != 5 {
		t.Errorf("expected cursor at 0:5, got %d:%d", ed.CursorY, ed.CursorX)
	}

	// 7. ToggleLineNumbers (defaults to false for classic Borland feel)
	if ed.ShowLineNums {
		t.Errorf("expected ShowLineNums initially false")
	}
	shown := ed.ToggleLineNumbers()
	if !shown || !ed.ShowLineNums {
		t.Errorf("expected ShowLineNums true after toggle")
	}
	shown = ed.ToggleLineNumbers()
	if shown || ed.ShowLineNums {
		t.Errorf("expected ShowLineNums false after second toggle")
	}
}

func TestFileBreakpointsIsolation(t *testing.T) {
	ed := NewEditor("fileA.go", 1)
	ed.Lines = []string{"line1", "line2", "line3"}

	// Set BP on fileA line 2
	ed.ToggleBreakpoint(2)
	if !ed.Breakpoints[2] {
		t.Fatalf("expected BP on line 2 in fileA")
	}

	// Switch to fileB (simulate load)
	ed.FilePath = "fileB.go"
	// LoadFile would do this:
	// But let's test LoadFile with real files or simulated logic:
	if len(ed.FileBreakpoints["fileA.go"]) == 0 {
		t.Errorf("expected fileA breakpoints saved in FileBreakpoints")
	}
}

func TestEditorFindNext(t *testing.T) {
	ed := NewEditor("test.go", 1)
	ed.Lines = []string{
		"package main",
		"import \"fmt\"",
		"func Hello() {",
		"    fmt.Println(\"hello world\")",
		"}",
	}
	ed.CursorX = 0
	ed.CursorY = 0

	// 1. Case-insensitive search for "hello" -> should match line 2 "Hello()"
	found := ed.FindNext("hello", false)
	if !found {
		t.Fatalf("expected to find 'hello' case-insensitive")
	}
	if ed.CursorY != 2 || ed.CursorX != 5 {
		t.Errorf("expected match at (2, 5), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// 2. Next search -> should match line 3 "hello world"
	found = ed.FindNext("hello", false)
	if !found {
		t.Fatalf("expected to find second 'hello'")
	}
	if ed.CursorY != 3 || ed.CursorX != 17 {
		t.Errorf("expected second match at (3, 17), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// 3. Case-sensitive search for "World" -> not found (in file it's "world")
	found = ed.FindNext("World", true)
	if found {
		t.Errorf("expected 'World' case-sensitive to fail")
	}

	// 4. Case-sensitive search for "world" -> found at line 3 col 23
	found = ed.FindNext("world", true)
	if !found || ed.CursorY != 3 || ed.CursorX != 23 {
		t.Errorf("expected 'world' match at (3, 23), got (%d, %d)", ed.CursorY, ed.CursorX)
	}
}

func TestEditorSelectionAndClipboard(t *testing.T) {
	ed := NewEditor("test.go", 1)
	ed.Lines = []string{
		"hello turbo go",
	}
	ed.CursorX = 6
	ed.CursorY = 0

	// Select "turbo" (from col 6 to 11)
	ed.StartSelection()
	ed.CursorX = 11
	ed.UpdateSelection()

	txt := ed.GetSelectedText()
	if txt != "turbo" {
		t.Fatalf("expected selected text 'turbo', got %q", txt)
	}

	// Test Copy
	ed.CopySelection()
	clip := GetClipboard()
	if clip != "turbo" {
		t.Errorf("expected clipboard 'turbo', got %q", clip)
	}

	// Test Cut
	ed.CutSelection()
	if ed.Lines[0] != "hello  go" {
		t.Errorf("expected line after cut 'hello  go', got %q", ed.Lines[0])
	}

	// Test Paste
	ed.PasteText("world")
	if ed.Lines[0] != "hello world go" {
		t.Errorf("expected line after paste 'hello world go', got %q", ed.Lines[0])
	}
}
