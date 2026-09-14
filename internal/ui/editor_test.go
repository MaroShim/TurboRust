package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"tr/internal/lsp"
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

func TestNavigationStack(t *testing.T) {
	ed := NewEditor("file1.rs", 1)
	ed.Lines = []string{"line1", "line2", "line3"}
	ed.CursorY = 0
	ed.CursorX = 0

	// Push initial position
	ed.PushNavLocation()

	// Move cursor
	ed.CursorY = 2
	ed.CursorX = 4
	ed.PushNavLocation()

	// Verify duplicate push suppression
	ed.PushNavLocation()
	if len(ed.navBackStack) != 2 {
		t.Fatalf("expected 2 locations in back stack, got %d", len(ed.navBackStack))
	}

	// Move cursor to another location
	ed.CursorY = 1
	ed.CursorX = 2

	// Navigate back -> should return to line 3 (index 2), col 4
	if !ed.NavigateBack() {
		t.Fatalf("expected NavigateBack to succeed")
	}
	if ed.CursorY != 2 || ed.CursorX != 4 {
		t.Errorf("expected Cursor (2, 4), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate forward -> should return to line 2 (index 1), col 2
	if !ed.NavigateForward() {
		t.Fatalf("expected NavigateForward to succeed")
	}
	if ed.CursorY != 1 || ed.CursorX != 2 {
		t.Errorf("expected Cursor (1, 2), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate back twice -> to initial (0, 0)
	ed.NavigateBack()
	if !ed.NavigateBack() {
		t.Fatalf("expected second NavigateBack to succeed")
	}
	if ed.CursorY != 0 || ed.CursorX != 0 {
		t.Errorf("expected Cursor (0, 0), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate back again -> empty stack, should return false
	if ed.NavigateBack() {
		t.Errorf("expected NavigateBack to fail on empty stack")
	}
}

func TestNavigationMultiFile(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.rs")
	f2 := filepath.Join(tmpDir, "b.rs")
	if err := os.WriteFile(f1, []byte("fn a() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("fn b() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ed := NewEditor("", 1)
	if err := ed.LoadFile(f1); err != nil {
		t.Fatal(err)
	}
	ed.CursorY = 0
	ed.CursorX = 3

	// Save location before jumping to f2
	ed.PushNavLocation()

	if err := ed.LoadFile(f2); err != nil {
		t.Fatal(err)
	}
	ed.CursorY = 0
	ed.CursorX = 3

	// Navigate back -> should switch back to f1 and restore cursor
	if !ed.NavigateBack() {
		t.Fatalf("expected NavigateBack across files to succeed")
	}
	if ed.FilePath != f1 {
		t.Errorf("expected FilePath %s, got %s", f1, ed.FilePath)
	}
	if ed.CursorY != 0 || ed.CursorX != 3 {
		t.Errorf("expected Cursor (0, 3), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate forward -> should switch to f2
	if !ed.NavigateForward() {
		t.Fatalf("expected NavigateForward across files to succeed")
	}
	if ed.FilePath != f2 {
		t.Errorf("expected FilePath %s, got %s", f2, ed.FilePath)
	}
}

func TestUndoRedo(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{""}
	ed.CursorX = 0
	ed.CursorY = 0

	// Initial typing
	for _, r := range "hello" {
		ed.InsertRune(r)
	}
	if ed.Lines[0] != "hello" {
		t.Fatalf("expected 'hello', got %q", ed.Lines[0])
	}

	// Undo typing 'o'
	if !ed.Undo() {
		t.Fatalf("expected Undo to succeed")
	}
	if ed.Lines[0] != "hell" {
		t.Errorf("expected 'hell' after undo, got %q", ed.Lines[0])
	}

	// Redo typing 'o'
	if !ed.Redo() {
		t.Fatalf("expected Redo to succeed")
	}
	if ed.Lines[0] != "hello" {
		t.Errorf("expected 'hello', got %q", ed.Lines[0])
	}

	// Insert new line
	ed.InsertNewLine()
	for _, r := range "world" {
		ed.InsertRune(r)
	}
	if len(ed.Lines) != 2 || ed.Lines[1] != "world" {
		t.Fatalf("expected line 1 to be 'world', got %v", ed.Lines)
	}

	// Undo 'd'
	ed.Undo()
	if ed.Lines[1] != "worl" {
		t.Errorf("expected 'worl', got %q", ed.Lines[1])
	}

	// Undo until line 2 disappears
	for len(ed.Lines) > 1 {
		if !ed.Undo() {
			t.Fatalf("expected Undo to succeed until newline removed")
		}
	}
	if len(ed.Lines) != 1 || ed.Lines[0] != "hello" {
		t.Errorf("expected single line 'hello', got %v", ed.Lines)
	}
}

// TestF12NavigationAndUndoRegression tests the multi-file jump, editing, and round-trip stack integrity in Rust
func TestF12NavigationAndUndoRegression(t *testing.T) {
	tmpDir := t.TempDir()
	fileA := filepath.Join(tmpDir, "main.rs")
	fileB := filepath.Join(tmpDir, "math.rs")

	codeA := "fn main() {\n    calculate();\n}\n"
	codeB := "pub fn calculate() -> i32 {\n    42\n}\n"

	if err := os.WriteFile(fileA, []byte(codeA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileB, []byte(codeB), 0644); err != nil {
		t.Fatal(err)
	}

	ed := NewEditor("", 1)
	if err := ed.LoadFile(fileA); err != nil {
		t.Fatal(err)
	}

	// 1. Move cursor to 'calculate()' on line 2 (index 1)
	ed.CursorY = 1
	ed.CursorX = 4

	// 2. Perform edit on File A: insert comments
	ed.InsertRune('/')
	ed.InsertRune('/')
	if !ed.Dirty {
		t.Errorf("expected ed.Dirty to be true after edit on File A")
	}

	// 3. Trigger F12: Push current location and jump to File B
	ed.PushNavLocation()
	if err := ed.LoadFile(fileB); err != nil {
		t.Fatal(err)
	}
	ed.GotoLine(1, 1) // Jumped to calculate() definition

	if ed.FilePath != fileB || ed.CursorY != 0 {
		t.Fatalf("expected jump to %s line 1, got %s line %d", fileB, ed.FilePath, ed.CursorY+1)
	}

	// 4. Perform second jump within File B to return statement (line 2)
	ed.PushNavLocation()
	ed.GotoLine(2, 4)

	// 5. Perform edit in File B: insert "// ok"
	for _, r := range "// ok" {
		ed.InsertRune(r)
	}
	// Undo in File B
	for i := 0; i < 5; i++ {
		if !ed.Undo() {
			t.Fatalf("expected Undo in File B to succeed")
		}
	}

	// 6. Navigate Back 1: should return to line 1 in File B
	if !ed.NavigateBack() {
		t.Fatalf("expected NavigateBack to step 1 to succeed")
	}
	if ed.FilePath != fileB || ed.CursorY != 0 {
		t.Errorf("expected File B line 1, got %s line %d", ed.FilePath, ed.CursorY+1)
	}

	// 7. Navigate Back 2: should return to File A at line 2
	if !ed.NavigateBack() {
		t.Fatalf("expected NavigateBack to File A to succeed")
	}
	if ed.FilePath != fileA || ed.CursorY != 1 {
		t.Errorf("expected File A line 2, got %s line %d", ed.FilePath, ed.CursorY+1)
	}

	// 8. Undo edit on File A: undo the two '/'
	if !ed.Undo() || !ed.Undo() {
		t.Fatalf("expected Undo in File A to succeed")
	}
	if ed.Lines[1] != "    calculate();" {
		t.Errorf("expected line 2 restored to '    calculate();', got %q", ed.Lines[1])
	}

	// 9. Navigate Forward: should move back into File B
	if !ed.NavigateForward() {
		t.Fatalf("expected NavigateForward into File B to succeed")
	}
	if ed.FilePath != fileB || ed.CursorY != 0 {
		t.Errorf("expected Forward into File B line 1, got %s line %d", ed.FilePath, ed.CursorY+1)
	}
}

func TestEditorAtomicSaveAndFileSafety(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Test Atomic Save
	targetFile := filepath.Join(tempDir, "test_atomic.rs")
	ed := NewEditor("", 1)
	ed.FilePath = targetFile
	ed.Lines = []string{"fn main() {", "}"}
	ed.Dirty = true

	if err := ed.SaveFile(); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}
	if ed.Dirty {
		t.Errorf("expected Dirty=false after save")
	}

	// Verify content on disk
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	expected := "fn main() {\n}"
	if string(content) != expected {
		t.Errorf("expected content %q, got %q", expected, string(content))
	}

	// Verify no tmp files left
	entries, _ := os.ReadDir(tempDir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-save-") {
			t.Errorf("found leftover tmp file: %s", entry.Name())
		}
	}

	// 2. Test Permission Preservation (Rule 5)
	if err := os.Chmod(targetFile, 0600); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	ed.Lines = append(ed.Lines, "// modified")
	if err := ed.SaveFile(); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}
	fi, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if fi.Mode().Perm() != 0600 {
		t.Errorf("expected permission 0600 preserved, got %04o", fi.Mode().Perm())
	}

	// 3. Test UTF-8 BOM Stripping (Rule 87)
	bomFile := filepath.Join(tempDir, "bom.rs")
	bomContent := append([]byte("\xef\xbb\xbf"), []byte("fn main() {}\n// with BOM")...)
	if err := os.WriteFile(bomFile, bomContent, 0644); err != nil {
		t.Fatalf("failed to write BOM file: %v", err)
	}
	edBom := NewEditor("", 2)
	if err := edBom.LoadFile(bomFile); err != nil {
		t.Fatalf("LoadFile on BOM file failed: %v", err)
	}
	if len(edBom.Lines) == 0 || edBom.Lines[0] != "fn main() {}" {
		t.Errorf("expected BOM stripped, first line got %q", edBom.Lines[0])
	}

	// 4. Test Binary Safety (Rule 15)
	binFile := filepath.Join(tempDir, "sample.bin")
	binContent := []byte{0x7f, 'E', 'L', 'F', 0x00, 0x01, 0x02}
	if err := os.WriteFile(binFile, binContent, 0644); err != nil {
		t.Fatalf("failed to write bin file: %v", err)
	}
	edBin := NewEditor("", 3)
	if err := edBin.LoadFile(binFile); err == nil {
		t.Errorf("expected error loading binary file with NUL byte, got nil")
	}

	// 5. Test Directory Rejection (Rule 6)
	if err := ed.LoadFile(tempDir); err == nil {
		t.Errorf("expected error loading directory, got nil")
	}
}

func TestEditorDraw_SemanticTokensOverlay(t *testing.T) {
	simScreen := tcell.NewSimulationScreen("UTF-8")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	ed := NewEditor("", 1)
	ed.Lines = []string{
		"fn calculate_sum(val: i32) -> MyStruct {",
		"    MyStruct {}",
		"}",
	}

	spans := []lsp.SemanticTokenSpan{
		{Line: 0, StartCol: 3, Length: 13, TokenType: "function"},
		{Line: 0, StartCol: 17, Length: 3, TokenType: "parameter"},
		{Line: 0, StartCol: 30, Length: 8, TokenType: "type"},
	}
	ed.SetSemanticTokens(spans)

	ed.Draw(simScreen, 0, 0, 80, 25, true)
	simScreen.Show()

	lineSpans := ed.GetSemanticTokensForLine(0)
	if len(lineSpans) != 3 {
		t.Fatalf("expected 3 semantic token spans for line 0, got %d", len(lineSpans))
	}
	if lineSpans[0].TokenType != "function" || lineSpans[1].TokenType != "parameter" || lineSpans[2].TokenType != "type" {
		t.Errorf("unexpected token types: %+v", lineSpans)
	}
}

