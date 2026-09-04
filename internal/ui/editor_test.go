package ui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestClipboardSystem(t *testing.T) {
	testStr := "TurboRustClipboardTest_12345"
	SetClipboard(testStr)
	got := GetClipboard()
	t.Logf("Set %q, Got %q", testStr, got)
	if got != testStr {
		t.Errorf("expected %q, got %q", testStr, got)
	}
}

func TestCutAndPasteWorkflow(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{"Hello World from TurboRust"}
	ed.CursorY = 0
	ed.CursorX = 6

	// Select "World"
	ed.StartSelection()
	ed.CursorX = 11
	ed.UpdateSelection()

	// Cut "World"
	if !ed.CutSelection() {
		t.Fatalf("CutSelection failed")
	}
	if ed.Lines[0] != "Hello  from TurboRust" {
		t.Errorf("expected 'Hello  from TurboRust', got %q", ed.Lines[0])
	}

	// Verify clipboard has "World"
	clip := GetClipboard()
	if clip != "World" {
		t.Errorf("expected clipboard 'World', got %q", clip)
	}

	// Move to end and Paste
	ed.CursorX = len([]rune(ed.Lines[0]))
	ed.PasteText(clip)

	expected := "Hello  from TurboRustWorld"
	if ed.Lines[0] != expected {
		t.Errorf("expected %q, got %q", expected, ed.Lines[0])
	}
}

func TestMultiLineCutAndPaste(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{
		"line 1: start",
		"line 2: middle content",
		"line 3: end of block",
		"line 4: unaffected",
	}

	// Select from line 0, col 8 ("start") to line 2, col 6 ("end of")
	ed.CursorY = 0
	ed.CursorX = 8
	ed.StartSelection()

	ed.CursorY = 2
	ed.CursorX = 6
	ed.UpdateSelection()

	selectedText := ed.GetSelectedText()
	expectedSelected := "start\nline 2: middle content\nline 3"
	if selectedText != expectedSelected {
		t.Fatalf("expected selected:\n%q\ngot:\n%q", expectedSelected, selectedText)
	}

	// Cut multi-line selection
	if !ed.CutSelection() {
		t.Fatalf("CutSelection failed")
	}

	if len(ed.Lines) != 2 {
		t.Fatalf("expected 2 lines remaining, got %d: %v", len(ed.Lines), ed.Lines)
	}
	if ed.Lines[0] != "line 1: : end of block" {
		t.Errorf("expected 'line 1: : end of block', got %q", ed.Lines[0])
	}
	if ed.Lines[1] != "line 4: unaffected" {
		t.Errorf("expected 'line 4: unaffected', got %q", ed.Lines[1])
	}

	// Paste the multi-line text at line 1, col 0
	ed.CursorY = 1
	ed.CursorX = 0
	ed.PasteText(selectedText)

	if len(ed.Lines) != 4 {
		t.Fatalf("expected 4 lines after paste, got %d: %v", len(ed.Lines), ed.Lines)
	}
	if ed.Lines[1] != "start" {
		t.Errorf("expected line 1 'start', got %q", ed.Lines[1])
	}
	if ed.Lines[2] != "line 2: middle content" {
		t.Errorf("expected line 2 'line 2: middle content', got %q", ed.Lines[2])
	}
	if ed.Lines[3] != "line 3line 4: unaffected" {
		t.Errorf("expected line 3 'line 3line 4: unaffected', got %q", ed.Lines[3])
	}
}

func TestSelectAllWorkflow(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{
		"fn main() {",
		"    println!(\"hello\");",
		"}",
	}

	ed.SelectAll()
	txt := ed.GetSelectedText()
	expected := "fn main() {\n    println!(\"hello\");\n}"
	if txt != expected {
		t.Fatalf("expected SelectAll to capture full buffer, got %q", txt)
	}

	// Delete all
	if !ed.DeleteSelection() {
		t.Fatalf("DeleteSelection on SelectAll failed")
	}
	if len(ed.Lines) != 1 || ed.Lines[0] != "" {
		t.Errorf("expected buffer cleared to single empty line, got %v", ed.Lines)
	}
	if ed.CursorX != 0 || ed.CursorY != 0 {
		t.Errorf("expected cursor at (0, 0), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Paste back
	ed.PasteText(txt)
	if strings.Join(ed.Lines, "\n") != expected {
		t.Errorf("expected buffer restored after paste, got %v", ed.Lines)
	}
}

func TestEditorBoundariesAndEdges(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{"abc", "def"}

	// 1. Backspace at (0, 0) should do nothing
	ed.CursorY = 0
	ed.CursorX = 0
	ed.Backspace()
	if len(ed.Lines) != 2 || ed.Lines[0] != "abc" {
		t.Errorf("backspace at (0, 0) should not corrupt lines")
	}

	// 2. Delete at end of line 0 should join line 1
	ed.CursorX = 3
	ed.Delete()
	if len(ed.Lines) != 1 || ed.Lines[0] != "abcdef" {
		t.Errorf("delete at EOL should join lines, got %v", ed.Lines)
	}

	// 3. Delete at EOF should do nothing
	ed.CursorX = len(ed.Lines[0])
	ed.Delete()
	if len(ed.Lines) != 1 || ed.Lines[0] != "abcdef" {
		t.Errorf("delete at EOF should do nothing, got %v", ed.Lines)
	}

	// 4. GotoLine out of range
	ed.GotoLine(-10, -5)
	if ed.CursorY != 0 || ed.CursorX != 0 {
		t.Errorf("expected clamp to (0, 0), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	ed.GotoLine(9999, 9999)
	if ed.CursorY != 0 || ed.CursorX != 6 {
		t.Errorf("expected clamp to (0, 6), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// 5. Paste empty string
	ed.PasteText("")
	if ed.Lines[0] != "abcdef" {
		t.Errorf("PasteText empty string should be no-op")
	}

	// 6. Paste with Windows CRLF (\r\n) line endings
	crlfText := "lineA\r\nlineB\r\nlineC"
	ed.GotoLine(1, 1)
	ed.Lines = []string{""}
	ed.PasteText(crlfText)
	if len(ed.Lines) != 3 {
		t.Fatalf("expected 3 lines after CRLF paste, got %d: %v", len(ed.Lines), ed.Lines)
	}
	if ed.Lines[0] != "lineA" || ed.Lines[1] != "lineB" || ed.Lines[2] != "lineC" {
		t.Errorf("CRLF not normalized cleanly: %v", ed.Lines)
	}
}

func TestMenuBarComprehensive(t *testing.T) {
	mb := NewMenuBar()

	// 1. Verify all 9 menus exist
	expectedTitles := []string{"File", "Edit", "Search", "Run", "Compile", "Debug", "Options", "Window", "Help"}
	if len(mb.Menus) != len(expectedTitles) {
		t.Fatalf("expected %d menus, got %d", len(expectedTitles), len(mb.Menus))
	}
	for i, exp := range expectedTitles {
		if mb.Menus[i].Title != exp {
			t.Errorf("menu %d expected title %q, got %q", i, exp, mb.Menus[i].Title)
		}
	}

	// 2. OpenMenu boundary checks
	mb.OpenMenu(-1)
	if mb.ActiveMenu != 0 {
		t.Errorf("expected clamp to 0 on negative index")
	}
	mb.OpenMenu(100)
	if mb.ActiveMenu != 0 {
		t.Errorf("expected clamp to keep previous valid index on overflow")
	}

	// 3. MoveRight wraps around
	mb.OpenMenu(len(mb.Menus) - 1) // Help (index 8)
	mb.MoveRight()
	if mb.ActiveMenu != 0 {
		t.Errorf("expected wrap around to File (0), got %d", mb.ActiveMenu)
	}

	// 4. MoveLeft wraps around
	mb.MoveLeft()
	if mb.ActiveMenu != len(mb.Menus)-1 {
		t.Errorf("expected wrap around to Help (8), got %d", mb.ActiveMenu)
	}

	// 5. File menu navigation skipping separators
	mb.OpenMenu(0) // File
	mb.ActiveItem = 3 // Save As
	mb.MoveDown()     // should skip item 4 (separator) and land on 5 (Exit)
	if mb.ActiveItem != 5 {
		t.Errorf("expected MoveDown to skip separator and reach index 5, got %d", mb.ActiveItem)
	}
	if mb.GetSelectedAction() != "app_exit" {
		t.Errorf("expected action 'app_exit', got %q", mb.GetSelectedAction())
	}

	mb.MoveUp() // should skip item 4 and land back on 3
	if mb.ActiveItem != 3 {
		t.Errorf("expected MoveUp to skip separator and reach index 3, got %d", mb.ActiveItem)
	}
	if mb.GetSelectedAction() != "file_save_as" {
		t.Errorf("expected action 'file_save_as', got %q", mb.GetSelectedAction())
	}
}

func TestStatusBarLifecycle(t *testing.T) {
	sb := NewStatusBar()
	if sb.Message != "" {
		t.Errorf("expected initial message to be empty")
	}

	// Set message
	sb.SetMessage("File saved successfully")
	if sb.Message != "File saved successfully" {
		t.Errorf("expected message to be set")
	}

	// Check expiry logic
	if time.Since(sb.MsgTime) > 1*time.Second {
		t.Errorf("fresh message should not be expired")
	}

	// Simulate expired message
	sb.MsgTime = time.Now().Add(-5 * time.Second)
	if time.Since(sb.MsgTime) <= 3*time.Second {
		t.Errorf("simulated 5s-old message should be expired")
	}
}

func TestEditorSaveAndLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "roundtrip_test.rs")

	ed1 := NewEditor("", 1)
	ed1.Lines = []string{
		"fn main() {",
		"    // 한글 주석 테스트 및 특수 기호: ╔═╗, 🚀, ©",
		"    let val = 42;",
		"}",
	}
	ed1.ToggleBreakpoint(3)

	err := ed1.SaveAs(filePath)
	if err != nil {
		t.Fatalf("SaveAs failed: %v", err)
	}

	// 1. Load into second editor to verify content
	ed2 := NewEditor("", 2)
	err = ed2.LoadFile(filePath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if len(ed2.Lines) != len(ed1.Lines) {
		t.Fatalf("expected %d lines, got %d", len(ed1.Lines), len(ed2.Lines))
	}
	for i := range ed1.Lines {
		if ed2.Lines[i] != ed1.Lines[i] {
			t.Errorf("line %d mismatch: expected %q, got %q", i, ed1.Lines[i], ed2.Lines[i])
		}
	}

	// 2. Reload into ed1 to verify breakpoint retention across file reloads
	err = ed1.LoadFile(filePath)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if !ed1.Breakpoints[3] {
		t.Errorf("expected breakpoint at line 3 to be restored after LoadFile")
	}
}

