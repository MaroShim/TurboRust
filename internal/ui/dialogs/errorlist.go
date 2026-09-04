package dialogs

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"tr/internal/compiler"
	"tr/internal/ui"
)

// ErrorListDialog displays compiler error list and allows jumping to line
type ErrorListDialog struct {
	Visible      bool
	Errors       []compiler.CompileError
	SelectedIndex int
	OnJump       func(err compiler.CompileError)
}

func NewErrorListDialog() *ErrorListDialog {
	return &ErrorListDialog{
		Visible: false,
	}
}

func (e *ErrorListDialog) Show(errs []compiler.CompileError, onJump func(err compiler.CompileError)) {
	e.Errors = errs
	e.SelectedIndex = 0
	e.OnJump = onJump
	e.Visible = true
}

func (e *ErrorListDialog) Hide() {
	e.Visible = false
}

func (e *ErrorListDialog) MoveUp() {
	if e.SelectedIndex > 0 {
		e.SelectedIndex--
	}
}

func (e *ErrorListDialog) MoveDown() {
	if e.SelectedIndex+1 < len(e.Errors) {
		e.SelectedIndex++
	}
}

func (e *ErrorListDialog) SelectCurrent() {
	if e.SelectedIndex >= 0 && e.SelectedIndex < len(e.Errors) {
		e.Visible = false
		if e.OnJump != nil {
			e.OnJump(e.Errors[e.SelectedIndex])
		}
	}
}

func (e *ErrorListDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !e.Visible {
		return
	}

	dialogW := screenW - 10
	if dialogW > 70 {
		dialogW = 70
	}
	if dialogW < 40 {
		dialogW = 40
	}
	dialogH := 10
	if len(e.Errors) < 6 {
		dialogH = len(e.Errors) + 5
	}
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	title := fmt.Sprintf("Compiler Messages (%d)", len(e.Errors))
	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, title)

	itemStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(tcell.ColorBlack)
	selStyle := tcell.StyleDefault.Background(tcell.ColorTeal).Foreground(tcell.ColorWhite).Bold(true)
	helpStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(tcell.ColorMaroon)

	viewH := dialogH - 3
	for i := 0; i < viewH && i < len(e.Errors); i++ {
		errItem := e.Errors[i]
		curStyle := itemStyle
		if i == e.SelectedIndex {
			curStyle = selStyle
		}

		lineStr := fmt.Sprintf(" %s:%d:%d: %s", errItem.File, errItem.Line, errItem.Column, errItem.Message)
		// Clear line inside dialog
		row := y + 1 + i
		for c := x + 1; c < x+dialogW-1; c++ {
			screen.SetContent(c, row, ' ', nil, curStyle)
		}

		col := x + 1
		for _, r := range lineStr {
			if col >= x+dialogW-1 {
				break
			}
			screen.SetContent(col, row, r, nil, curStyle)
			col += runewidth.RuneWidth(r)
		}
	}

	// Bottom guidance
	guidance := "[ Enter: Jump to Line   Esc: Close ]"
	gx := x + (dialogW-len(guidance))/2
	for i, r := range guidance {
		screen.SetContent(gx+i, y+dialogH-1, r, nil, helpStyle)
	}
}
