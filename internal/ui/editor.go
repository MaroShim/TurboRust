package ui

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"tr/internal/syntax"
)

// Editor holds the state of a code editing buffer
type Editor struct {
	Lines        []string
	CursorX      int // Visual/runic column in current line
	CursorY      int // Line index
	ScrollX      int
	ScrollY      int
	FilePath     string
	FileName     string
	Dirty        bool
	ShowLineNums bool
	WindowNumber int
	TabWidth     int // Default 4

	// Breakpoint tracking
	Breakpoints     map[int]bool            // 1-based line number -> is breakpoint for CURRENT file
	FileBreakpoints map[string]map[int]bool // file path -> breakpoints map
	CurrentIP       int                     // 1-based current execution instruction pointer (for debug)

	// Search tracking & visual match highlight
	LastFindQuery     string
	LastCaseSensitive bool
	HighlightLine     int // 0-based, -1 if no highlight
	HighlightStartCol int
	HighlightEndCol   int

	// Selection tracking
	SelectActive bool
	SelectStartY int
	SelectStartX int
	SelectEndY   int
	SelectEndX   int
}

func ExpandTabs(s string, tabWidth int) string {
	if tabWidth <= 0 {
		tabWidth = 4
	}
	var b strings.Builder
	col := 0
	for _, r := range s {
		if r == '\t' {
			spaces := tabWidth - (col % tabWidth)
			for i := 0; i < spaces; i++ {
				b.WriteRune(' ')
			}
			col += spaces
		} else {
			b.WriteRune(r)
			col++
		}
	}
	return b.String()
}

func NewEditor(filePath string, windowNum int) *Editor {
	ed := &Editor{
		Lines:             []string{""},
		CursorX:           0,
		CursorY:           0,
		FilePath:          filePath,
		FileName:          "NONAME00.RS",
		ShowLineNums:      false,
		WindowNumber:      windowNum,
		TabWidth:          4,
		Breakpoints:       make(map[int]bool),
		FileBreakpoints:   make(map[string]map[int]bool),
		HighlightLine:     -1,
		HighlightStartCol: 0,
		HighlightEndCol:   0,
	}

	if filePath != "" {
		ed.LoadFile(filePath)
	} else {
		// Sample starter retro template for Rust
		ed.Lines = []string{
			"fn main() {",
			"    // Welcome to Turbo Rust!",
			"    // Press Ctrl+F9 to Run, Alt+F9 to Compile",
			"    // Press Alt+F5 to see the User Screen",
			"    println!(\"Hello, Turbo Rust World!\");",
			"}",
		}
	}
	return ed
}

func (e *Editor) LoadFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 1. Save current file's breakpoints before switching
	if e.FilePath != "" && e.FileBreakpoints != nil {
		curKey := filepath.Clean(e.FilePath)
		if len(e.Breakpoints) > 0 {
			saved := make(map[int]bool)
			for k, v := range e.Breakpoints {
				if v {
					saved[k] = true
				}
			}
			e.FileBreakpoints[curKey] = saved
		} else {
			delete(e.FileBreakpoints, curKey)
		}
	}

	rawLines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	if len(rawLines) == 0 {
		rawLines = []string{""}
	}
	lines := make([]string, len(rawLines))
	for i, l := range rawLines {
		lines[i] = ExpandTabs(l, e.TabWidth)
	}
	e.Lines = lines
	e.FilePath = path
	e.FileName = filepath.Base(path)
	e.Dirty = false
	e.CursorX = 0
	e.CursorY = 0
	e.ScrollX = 0
	e.ScrollY = 0
	e.CurrentIP = 0

	// 2. Restore or init breakpoints for the new file
	if e.FileBreakpoints == nil {
		e.FileBreakpoints = make(map[string]map[int]bool)
	}
	newKey := filepath.Clean(path)
	if bps, ok := e.FileBreakpoints[newKey]; ok {
		e.Breakpoints = make(map[int]bool)
		for k, v := range bps {
			if v {
				e.Breakpoints[k] = true
			}
		}
	} else {
		e.Breakpoints = make(map[int]bool)
	}

	return nil
}

func (e *Editor) SaveFile() error {
	if e.FilePath == "" {
		e.FilePath = "main.go"
		e.FileName = "main.go"
	}
	content := strings.Join(e.Lines, "\n")
	err := os.WriteFile(e.FilePath, []byte(content), 0644)
	if err == nil {
		e.Dirty = false
	}
	return err
}

func (e *Editor) SaveAs(path string) error {
	e.FilePath = path
	e.FileName = filepath.Base(path)
	return e.SaveFile()
}

func (e *Editor) ToggleBreakpoint(line int) bool {
	if e.Breakpoints == nil {
		e.Breakpoints = make(map[int]bool)
	}
	isSet := false
	if e.Breakpoints[line] {
		delete(e.Breakpoints, line)
		isSet = false
	} else {
		e.Breakpoints[line] = true
		isSet = true
	}

	// Synchronize with FileBreakpoints
	if e.FilePath != "" {
		if e.FileBreakpoints == nil {
			e.FileBreakpoints = make(map[string]map[int]bool)
		}
		curKey := filepath.Clean(e.FilePath)
		if len(e.Breakpoints) > 0 {
			saved := make(map[int]bool)
			for k, v := range e.Breakpoints {
				if v {
					saved[k] = true
				}
			}
			e.FileBreakpoints[curKey] = saved
		} else {
			delete(e.FileBreakpoints, curKey)
		}
	}

	return isSet
}

func (e *Editor) ToggleLineNumbers() bool {
	e.ShowLineNums = !e.ShowLineNums
	return e.ShowLineNums
}

func (e *Editor) SetCurrentIP(line int) {
	e.CurrentIP = line
	if line > 0 {
		e.GotoLine(line, 1)
	}
}

func (e *Editor) GotoLine(line, col int) {
	if line < 1 {
		line = 1
	}
	if line > len(e.Lines) {
		line = len(e.Lines)
	}
	e.CursorY = line - 1

	lineLen := len([]rune(e.Lines[e.CursorY]))
	if col < 1 {
		col = 1
	}
	if col-1 > lineLen {
		e.CursorX = lineLen
	} else {
		e.CursorX = col - 1
	}
}

func (e *Editor) InsertRune(ch rune) {
	e.ClearHighlight()
	if e.SelectActive {
		e.DeleteSelection()
	}
	if ch == '\t' {
		e.InsertTab()
		return
	}
	if e.CursorY >= len(e.Lines) {
		e.Lines = append(e.Lines, "")
	}
	runes := []rune(e.Lines[e.CursorY])
	if e.CursorX > len(runes) {
		e.CursorX = len(runes)
	}

	newLine := make([]rune, 0, len(runes)+1)
	newLine = append(newLine, runes[:e.CursorX]...)
	newLine = append(newLine, ch)
	newLine = append(newLine, runes[e.CursorX:]...)

	e.Lines[e.CursorY] = string(newLine)
	e.CursorX++
	e.Dirty = true
}

func (e *Editor) InsertTab() {
	if e.SelectActive {
		e.DeleteSelection()
	}
	tabW := e.TabWidth
	if tabW <= 0 {
		tabW = 4
	}
	spaces := tabW - (e.CursorX % tabW)
	for i := 0; i < spaces; i++ {
		e.InsertRune(' ')
	}
}

func (e *Editor) InsertNewLine() {
	if e.SelectActive {
		e.DeleteSelection()
	}
	if e.CursorY >= len(e.Lines) {
		e.Lines = append(e.Lines, "")
	}
	runes := []rune(e.Lines[e.CursorY])
	if e.CursorX > len(runes) {
		e.CursorX = len(runes)
	}

	// Calculate auto-indent from current line
	indent := ""
	for _, r := range runes {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}

	left := string(runes[:e.CursorX])
	right := indent + string(runes[e.CursorX:])

	e.Lines[e.CursorY] = left
	newLines := make([]string, 0, len(e.Lines)+1)
	newLines = append(newLines, e.Lines[:e.CursorY+1]...)
	newLines = append(newLines, right)
	newLines = append(newLines, e.Lines[e.CursorY+1:]...)

	e.Lines = newLines
	e.CursorY++
	e.CursorX = len([]rune(indent))
	e.Dirty = true
}

func (e *Editor) Backspace() {
	e.ClearHighlight()
	if e.SelectActive {
		e.DeleteSelection()
		return
	}
	if e.CursorX > 0 {
		runes := []rune(e.Lines[e.CursorY])

		// Smart backspace: if only spaces precede cursor, delete up to TabWidth spaces
		tabW := e.TabWidth
		if tabW <= 0 {
			tabW = 4
		}
		allSpaces := true
		for i := 0; i < e.CursorX; i++ {
			if runes[i] != ' ' {
				allSpaces = false
				break
			}
		}

		spacesToDelete := 1
		if allSpaces {
			rem := e.CursorX % tabW
			if rem == 0 {
				spacesToDelete = tabW
			} else {
				spacesToDelete = rem
			}
			if spacesToDelete > e.CursorX {
				spacesToDelete = e.CursorX
			}
		}

		newLine := make([]rune, 0, len(runes)-spacesToDelete)
		newLine = append(newLine, runes[:e.CursorX-spacesToDelete]...)
		newLine = append(newLine, runes[e.CursorX:]...)
		e.Lines[e.CursorY] = string(newLine)
		e.CursorX -= spacesToDelete
		e.Dirty = true
	} else if e.CursorY > 0 {
		// Join with previous line
		prevRunes := []rune(e.Lines[e.CursorY-1])
		currRunes := []rune(e.Lines[e.CursorY])

		e.CursorX = len(prevRunes)
		e.Lines[e.CursorY-1] = string(append(prevRunes, currRunes...))

		// Remove current line
		e.Lines = append(e.Lines[:e.CursorY], e.Lines[e.CursorY+1:]...)
		e.CursorY--
		e.Dirty = true
	}
}

func (e *Editor) Delete() {
	e.ClearHighlight()
	if e.SelectActive {
		e.DeleteSelection()
		return
	}
	if e.CursorY >= len(e.Lines) {
		return
	}
	runes := []rune(e.Lines[e.CursorY])
	if e.CursorX < len(runes) {
		newLine := make([]rune, 0, len(runes)-1)
		newLine = append(newLine, runes[:e.CursorX]...)
		newLine = append(newLine, runes[e.CursorX+1:]...)
		e.Lines[e.CursorY] = string(newLine)
		e.Dirty = true
	} else if e.CursorY+1 < len(e.Lines) {
		// Join next line
		nextRunes := []rune(e.Lines[e.CursorY+1])
		e.Lines[e.CursorY] = string(append(runes, nextRunes...))
		e.Lines = append(e.Lines[:e.CursorY+1], e.Lines[e.CursorY+2:]...)
		e.Dirty = true
	}
}

func (e *Editor) MoveLeft() {
	e.ClearHighlight()
	if e.CursorX > 0 {
		e.CursorX--
	} else if e.CursorY > 0 {
		e.CursorY--
		e.CursorX = len([]rune(e.Lines[e.CursorY]))
	}
}

func (e *Editor) MoveRight() {
	e.ClearHighlight()
	runes := []rune(e.Lines[e.CursorY])
	if e.CursorX < len(runes) {
		e.CursorX++
	} else if e.CursorY+1 < len(e.Lines) {
		e.CursorY++
		e.CursorX = 0
	}
}

func (e *Editor) MoveUp() {
	e.ClearHighlight()
	if e.CursorY > 0 {
		e.CursorY--
		lineLen := len([]rune(e.Lines[e.CursorY]))
		if e.CursorX > lineLen {
			e.CursorX = lineLen
		}
	}
}

func (e *Editor) MoveDown() {
	e.ClearHighlight()
	if e.CursorY+1 < len(e.Lines) {
		e.CursorY++
		lineLen := len([]rune(e.Lines[e.CursorY]))
		if e.CursorX > lineLen {
			e.CursorX = lineLen
		}
	}
}

func (e *Editor) MoveHome() {
	e.CursorX = 0
}

func (e *Editor) MoveEnd() {
	if e.CursorY < len(e.Lines) {
		e.CursorX = len([]rune(e.Lines[e.CursorY]))
	}
}

func (e *Editor) PageUp(height int) {
	e.CursorY -= height
	if e.CursorY < 0 {
		e.CursorY = 0
	}
	lineLen := len([]rune(e.Lines[e.CursorY]))
	if e.CursorX > lineLen {
		e.CursorX = lineLen
	}
}

func (e *Editor) PageDown(height int) {
	e.CursorY += height
	if e.CursorY >= len(e.Lines) {
		e.CursorY = len(e.Lines) - 1
	}
	lineLen := len([]rune(e.Lines[e.CursorY]))
	if e.CursorX > lineLen {
		e.CursorX = lineLen
	}
}

// AdjustScroll ensures the cursor is visible within view width/height
func (e *Editor) AdjustScroll(width, height int) {
	if e.CursorY < e.ScrollY {
		e.ScrollY = e.CursorY
	}
	if e.CursorY >= e.ScrollY+height {
		e.ScrollY = e.CursorY - height + 1
	}

	if e.CursorX < e.ScrollX {
		e.ScrollX = e.CursorX
	}
	if e.CursorX >= e.ScrollX+width {
		e.ScrollX = e.CursorX - width + 1
	}
}

// Draw renders the editor content inside the given rectangle
func (e *Editor) Draw(screen tcell.Screen, x, y, width, height int, focused bool) {
	lineNumWidth := 0
	if e.ShowLineNums {
		lineNumWidth = 5 // " 123 "
	}
	codeWidth := width - lineNumWidth
	if codeWidth < 1 {
		codeWidth = 1
	}

	e.AdjustScroll(codeWidth, height)

	baseStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorFg)
	lineNumStyle := tcell.StyleDefault.Background(ColorEditorLineNumBg).Foreground(ColorEditorLineNumFg)

	inBlockComment := false

	for row := 0; row < height; row++ {
		lineIdx := e.ScrollY + row
		screenY := y + row

		lineNum1 := lineIdx + 1
		hasBP := false
		isIP := false
		if lineIdx < len(e.Lines) {
			hasBP = e.Breakpoints[lineNum1]
			isIP = (e.CurrentIP == lineNum1)
		}

		// Determine unified line background & text style
		lineBaseStyle := baseStyle
		if isIP {
			// B Style (Turbo Debugger / VC++ 6.0 Standard): Full Yellow Bar with Black Text
			lineBaseStyle = tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true)
		} else if hasBP {
			// Borland Classic: Full Red Bar with White Text
			lineBaseStyle = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)
		}

		// 1. Draw line numbers & breakpoint gutter
		if e.ShowLineNums {
			if lineIdx < len(e.Lines) {
				var gutterStyle tcell.Style
				var mark rune = ' '

				if isIP {
					gutterStyle = lineBaseStyle
					mark = RuneCurrentLine
				} else if hasBP {
					gutterStyle = lineBaseStyle
					mark = RuneBreakpoint
				} else {
					gutterStyle = lineNumStyle
				}

				// Format: " 123 " or "●123 " / "►123 "
				numStr := formatLineNum(lineNum1, lineNumWidth-1)
				screen.SetContent(x, screenY, mark, nil, gutterStyle)
				for i, r := range numStr {
					screen.SetContent(x+1+i, screenY, r, nil, gutterStyle)
				}
			} else {
				// Empty line after EOF
				for col := 0; col < lineNumWidth; col++ {
					screen.SetContent(x+col, screenY, ' ', nil, lineNumStyle)
				}
			}
		}

		// 2. Draw Code with Syntax Highlighting
		codeStartX := x + lineNumWidth
		if lineIdx < len(e.Lines) {
			lineText := e.Lines[lineIdx]
			tokens := syntax.HighlightLine(lineText, lineBaseStyle, &inBlockComment)

			screenX := codeStartX
			tabW := e.TabWidth
			if tabW <= 0 {
				tabW = 4
			}
			for col := e.ScrollX; col < len(tokens) && screenX < x+width; col++ {
				tok := tokens[col]
				tokStyle := tok.Style

				if isIP {
					// On Yellow bar: High-contrast Dark text
					tokStyle = tokStyle.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true)
				} else if hasBP {
					// On Red bar: Crisp White text
					tokStyle = tokStyle.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)
				} else if e.IsSelected(lineIdx, col) {
					// Classic Borland Block Selection: Inverted Crisp LightCyan block
					tokStyle = tcell.StyleDefault.Background(tcell.ColorLightCyan).Foreground(tcell.ColorBlack).Bold(true)
				} else if lineIdx == e.HighlightLine && col >= e.HighlightStartCol && col < e.HighlightEndCol {
					// Classic Borland Search Match Highlight: Crisp Light Cyan block with Black text
					tokStyle = tcell.StyleDefault.Background(tcell.ColorLightCyan).Foreground(tcell.ColorBlack).Bold(true)
				}

				if tok.Char == '\t' {
					relCol := screenX - codeStartX
					spaces := tabW - (relCol % tabW)
					for s := 0; s < spaces && screenX < x+width; s++ {
						screen.SetContent(screenX, screenY, ' ', nil, tokStyle)
						screenX++
					}
				} else {
					screen.SetContent(screenX, screenY, tok.Char, nil, tokStyle)
					screenX += runewidth.RuneWidth(tok.Char)
				}
			}
			// Fill rest of the line with lineBaseStyle (stretches yellow or red bar across full window width)
			for screenX < x+width {
				screen.SetContent(screenX, screenY, ' ', nil, lineBaseStyle)
				screenX++
			}
		} else {
			// Below EOF: classic tilde or empty space
			screen.SetContent(codeStartX, screenY, '~', nil, baseStyle.Foreground(ColorEditorLineNumFg))
			for col := codeStartX + 1; col < x+width; col++ {
				screen.SetContent(col, screenY, ' ', nil, baseStyle)
			}
		}
	}

	// 3. Set terminal hardware cursor if focused
	if focused {
		cursorScreenX := x + lineNumWidth + (e.CursorX - e.ScrollX)
		cursorScreenY := y + (e.CursorY - e.ScrollY)
		if cursorScreenX >= x+lineNumWidth && cursorScreenX < x+width &&
			cursorScreenY >= y && cursorScreenY < y+height {
			screen.ShowCursor(cursorScreenX, cursorScreenY)
		} else {
			screen.HideCursor()
		}
	}
}

func formatLineNum(num, width int) string {
	s := ""
	temp := num
	for temp > 0 {
		s = string(rune('0'+(temp%10))) + s
		temp /= 10
	}
	for len(s) < width {
		s = " " + s
	}
	return s
}

// FindNext searches forward from the current cursor position for the query string.
// If found, it moves cursor to the match, adjusts scroll, and returns true.
func (e *Editor) FindNext(query string, caseSensitive bool) bool {
	if query == "" || len(e.Lines) == 0 {
		return false
	}
	e.LastFindQuery = query
	e.LastCaseSensitive = caseSensitive

	matchAt := func(line string, idx int) bool {
		runes := []rune(line)
		qRunes := []rune(query)
		if idx+len(qRunes) > len(runes) {
			return false
		}
		for i := 0; i < len(qRunes); i++ {
			r1 := runes[idx+i]
			r2 := qRunes[i]
			if !caseSensitive {
				r1 = unicode.ToLower(r1)
				r2 = unicode.ToLower(r2)
			}
			if r1 != r2 {
				return false
			}
		}
		return true
	}

	startY := e.CursorY
	startX := e.CursorX + 1 // Advance 1 rune so repeated FindNext moves forward
	totalLines := len(e.Lines)

	// 1. Search from (startY, startX) to the end of document
	for count := 0; count < totalLines; count++ {
		curY := (startY + count) % totalLines
		line := e.Lines[curY]
		lineRunes := []rune(line)
		qLen := len([]rune(query))

		searchFromX := 0
		if count == 0 {
			searchFromX = startX
		}

		for x := searchFromX; x <= len(lineRunes)-qLen; x++ {
			if matchAt(line, x) {
				e.CursorY = curY
				e.CursorX = x
				e.HighlightLine = curY
				e.HighlightStartCol = x
				e.HighlightEndCol = x + qLen
				if e.CursorY < e.ScrollY || e.CursorY >= e.ScrollY+15 {
					e.ScrollY = e.CursorY - 5
					if e.ScrollY < 0 {
						e.ScrollY = 0
					}
				}
				if e.CursorX < e.ScrollX || e.CursorX >= e.ScrollX+50 {
					e.ScrollX = e.CursorX - 10
					if e.ScrollX < 0 {
						e.ScrollX = 0
					}
				}
				return true
			}
		}
	}

	// 2. Wrap around check on startY from column 0 up to startX
	if startX > 0 {
		line := e.Lines[startY]
		lineRunes := []rune(line)
		qLen := len([]rune(query))
		for x := 0; x < startX && x <= len(lineRunes)-qLen; x++ {
			if matchAt(line, x) {
				e.CursorY = startY
				e.CursorX = x
				e.HighlightLine = startY
				e.HighlightStartCol = x
				e.HighlightEndCol = x + qLen
				if e.CursorY < e.ScrollY || e.CursorY >= e.ScrollY+15 {
					e.ScrollY = e.CursorY - 5
					if e.ScrollY < 0 {
						e.ScrollY = 0
					}
				}
				if e.CursorX < e.ScrollX || e.CursorX >= e.ScrollX+50 {
					e.ScrollX = e.CursorX - 10
					if e.ScrollX < 0 {
						e.ScrollX = 0
					}
				}
				return true
			}
		}
	}

	e.HighlightLine = -1
	return false
}

// ClearHighlight removes any active search match highlight
func (e *Editor) ClearHighlight() {
	e.HighlightLine = -1
}

// ClearSelection cancels any active block selection
func (e *Editor) ClearSelection() {
	e.SelectActive = false
}

// StartSelection initializes block selection from current cursor
func (e *Editor) StartSelection() {
	if !e.SelectActive {
		e.SelectActive = true
		e.SelectStartY = e.CursorY
		e.SelectStartX = e.CursorX
	}
	e.SelectEndY = e.CursorY
	e.SelectEndX = e.CursorX
}

// UpdateSelection updates the moving end-point of block selection
func (e *Editor) UpdateSelection() {
	e.SelectEndY = e.CursorY
	e.SelectEndX = e.CursorX
}

// GetNormalizedSelection returns (startY, startX, endY, endX, ok)
func (e *Editor) GetNormalizedSelection() (int, int, int, int, bool) {
	if !e.SelectActive {
		return 0, 0, 0, 0, false
	}
	sy, sx := e.SelectStartY, e.SelectStartX
	ey, ex := e.SelectEndY, e.SelectEndX

	if sy > ey || (sy == ey && sx > ex) {
		sy, ey = ey, sy
		sx, ex = ex, sx
	}
	if sy == ey && sx == ex {
		return 0, 0, 0, 0, false
	}
	return sy, sx, ey, ex, true
}

// IsSelected checks if the character at (lineIdx, col) is within active selection
func (e *Editor) IsSelected(lineIdx, col int) bool {
	sy, sx, ey, ex, ok := e.GetNormalizedSelection()
	if !ok {
		return false
	}
	if lineIdx < sy || lineIdx > ey {
		return false
	}
	if lineIdx == sy && lineIdx == ey {
		return col >= sx && col < ex
	}
	if lineIdx == sy {
		return col >= sx
	}
	if lineIdx == ey {
		return col < ex
	}
	return true
}

// GetSelectedText returns the string currently highlighted by selection
func (e *Editor) GetSelectedText() string {
	sy, sx, ey, ex, ok := e.GetNormalizedSelection()
	if !ok {
		return ""
	}
	if sy == ey {
		runes := []rune(e.Lines[sy])
		if sx > len(runes) {
			sx = len(runes)
		}
		if ex > len(runes) {
			ex = len(runes)
		}
		return string(runes[sx:ex])
	}

	var parts []string
	r0 := []rune(e.Lines[sy])
	if sx > len(r0) {
		sx = len(r0)
	}
	parts = append(parts, string(r0[sx:]))

	for y := sy + 1; y < ey; y++ {
		parts = append(parts, e.Lines[y])
	}

	rN := []rune(e.Lines[ey])
	if ex > len(rN) {
		ex = len(rN)
	}
	parts = append(parts, string(rN[:ex]))

	return strings.Join(parts, "\n")
}

// DeleteSelection removes the selected block and places cursor at start
func (e *Editor) DeleteSelection() bool {
	sy, sx, ey, ex, ok := e.GetNormalizedSelection()
	if !ok {
		return false
	}
	if sy == ey {
		r := []rune(e.Lines[sy])
		if sx > len(r) {
			sx = len(r)
		}
		if ex > len(r) {
			ex = len(r)
		}
		e.Lines[sy] = string(append(r[:sx], r[ex:]...))
	} else {
		r0 := []rune(e.Lines[sy])
		rN := []rune(e.Lines[ey])
		if sx > len(r0) {
			sx = len(r0)
		}
		if ex > len(rN) {
			ex = len(rN)
		}
		merged := string(append(r0[:sx], rN[ex:]...))
		newLines := make([]string, 0, len(e.Lines)-(ey-sy))
		newLines = append(newLines, e.Lines[:sy]...)
		newLines = append(newLines, merged)
		if ey+1 < len(e.Lines) {
			newLines = append(newLines, e.Lines[ey+1:]...)
		}
		e.Lines = newLines
	}
	e.CursorY = sy
	e.CursorX = sx
	e.ClearSelection()
	e.Dirty = true
	return true
}

// CopySelection copies selected text to clipboard
func (e *Editor) CopySelection() bool {
	txt := e.GetSelectedText()
	if txt != "" {
		SetClipboard(txt)
		return true
	}
	return false
}

// CutSelection copies selected text to clipboard and deletes it
func (e *Editor) CutSelection() bool {
	if e.CopySelection() {
		return e.DeleteSelection()
	}
	return false
}

// PasteText inserts text at current cursor position (replacing any selection)
func (e *Editor) PasteText(text string) {
	if text == "" {
		return
	}
	if e.SelectActive {
		e.DeleteSelection()
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 1 {
		runes := []rune(e.Lines[e.CursorY])
		if e.CursorX > len(runes) {
			e.CursorX = len(runes)
		}
		newLine := make([]rune, 0, len(runes)+len([]rune(text)))
		newLine = append(newLine, runes[:e.CursorX]...)
		newLine = append(newLine, []rune(text)...)
		newLine = append(newLine, runes[e.CursorX:]...)
		e.Lines[e.CursorY] = string(newLine)
		e.CursorX += len([]rune(text))
	} else {
		currRunes := []rune(e.Lines[e.CursorY])
		if e.CursorX > len(currRunes) {
			e.CursorX = len(currRunes)
		}
		left := string(currRunes[:e.CursorX])
		right := string(currRunes[e.CursorX:])

		firstLine := left + lines[0]
		lastLine := lines[len(lines)-1] + right

		newLines := make([]string, 0, len(e.Lines)+len(lines)-1)
		newLines = append(newLines, e.Lines[:e.CursorY]...)
		newLines = append(newLines, firstLine)
		for i := 1; i < len(lines)-1; i++ {
			newLines = append(newLines, lines[i])
		}
		newLines = append(newLines, lastLine)
		if e.CursorY+1 < len(e.Lines) {
			newLines = append(newLines, e.Lines[e.CursorY+1:]...)
		}

		e.Lines = newLines
		e.CursorY += len(lines) - 1
		e.CursorX = len([]rune(lines[len(lines)-1]))
	}
	e.Dirty = true
}
