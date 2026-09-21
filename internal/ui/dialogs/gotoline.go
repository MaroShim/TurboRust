package dialogs

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/MaroShim/TurboRust/internal/ui"
)

// GotoLineDialog allows jumping directly to a line number
type GotoLineDialog struct {
	Visible    bool
	LineText   string
	OnJump     func(line int)
	TotalLines int
}

func NewGotoLineDialog() *GotoLineDialog {
	return &GotoLineDialog{
		Visible: false,
	}
}

func (g *GotoLineDialog) Show(currentLine, totalLines int, onJump func(line int)) {
	g.LineText = fmt.Sprintf("%d", currentLine)
	g.TotalLines = totalLines
	g.OnJump = onJump
	g.Visible = true
}

func (g *GotoLineDialog) Hide() {
	g.Visible = false
}

func (g *GotoLineDialog) InsertRune(ch rune) {
	if ch >= '0' && ch <= '9' && len(g.LineText) < 6 {
		g.LineText += string(ch)
	}
}

func (g *GotoLineDialog) Backspace() {
	runes := []rune(g.LineText)
	if len(runes) > 0 {
		g.LineText = string(runes[:len(runes)-1])
	}
}

func (g *GotoLineDialog) Confirm() {
	val, err := strconv.Atoi(g.LineText)
	if err == nil && val > 0 {
		g.Visible = false
		if g.OnJump != nil {
			g.OnJump(val)
		}
	} else {
		g.Visible = false
	}
}

func (g *GotoLineDialog) IsVisible() bool {
	return g.Visible
}

func (g *GotoLineDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !g.Visible {
		return
	}

	dialogW := 36
	dialogH := 9
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "Go to Line")

	textStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)
	hintStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(tcell.ColorDarkGray)

	// Label: "Enter line number:"
	label := fmt.Sprintf("Line number (1..%d):", g.TotalLines)
	if len(label) > dialogW-6 {
		label = "Line number:"
	}
	for i, r := range label {
		screen.SetContent(x+3+i, y+2, r, nil, textStyle)
	}

	// Input box
	boxW := dialogW - 6
	ui.DrawInputField(screen, x+3, y+3, boxW, g.LineText, true)

	// Hint
	hint := "[ Enter: OK   Esc: Cancel ]"
	hx := x + (dialogW-len(hint))/2
	for i, r := range hint {
		screen.SetContent(hx+i, y+dialogH-4, r, nil, hintStyle)
	}

	// Buttons
	ui.DrawButton(screen, x+4, y+dialogH-2, 10, "OK", true)
	ui.DrawButton(screen, x+18, y+dialogH-2, 10, "Cancel", false)
}

// HandleMouse processes mouse events when GotoLineDialog is open.
func (g *GotoLineDialog) HandleMouse(mx, my int, btn tcell.ButtonMask, screenW, screenH int) bool {
	if !g.Visible {
		return false
	}
	dialogW := 36
	dialogH := 9
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	if btn&tcell.Button1 != 0 {
		// OK button: x+4, y+dialogH-2, width 10
		if my == y+dialogH-2 && mx >= x+4 && mx < x+4+10 {
			g.Confirm()
			return true
		}
		// Cancel button: x+18, y+dialogH-2, width 10
		if my == y+dialogH-2 && mx >= x+18 && mx < x+18+10 {
			g.Hide()
			return true
		}

		if mx >= x && mx < x+dialogW && my >= y && my < y+dialogH {
			return true
		}

		// Click outside: cancel
		g.Hide()
		return true
	}
	return true
}

