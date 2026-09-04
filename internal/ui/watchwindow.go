package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"tr/internal/debugger"
)

// WatchWindow displays inspected variables and debugger status at the bottom of the IDE
type WatchWindow struct {
	Visible      bool
	Variables    []debugger.Variable
	ScrollY      int
	StatusText   string
	WindowNumber int
}

func NewWatchWindow(winNum int) *WatchWindow {
	return &WatchWindow{
		Visible:      false,
		WindowNumber: winNum,
		StatusText:   "No active debug session.",
	}
}

func (w *WatchWindow) SetState(st debugger.DebugState) {
	w.Variables = st.LocalVars
	if st.Exited {
		w.StatusText = fmt.Sprintf("Program exited (code %d). [Alt+F5: User Screen]", st.ExitCode)
	} else if st.CurrentFile != "" && st.CurrentLine > 0 {
		base := st.CurrentFile
		for i := len(base) - 1; i >= 0; i-- {
			if base[i] == '/' || base[i] == '\\' {
				base = base[i+1:]
				break
			}
		}
		w.StatusText = fmt.Sprintf("Paused at %s:%d in %s", base, st.CurrentLine, st.CurrentFunc)
	} else if st.Active {
		w.StatusText = "Debugging session active (running)..."
	} else {
		w.StatusText = "No active debug session."
	}
}

func (w *WatchWindow) ScrollUp() {
	if w.ScrollY > 0 {
		w.ScrollY--
	}
}

func (w *WatchWindow) ScrollDown(height int) {
	if w.ScrollY+height < len(w.Variables) {
		w.ScrollY++
	}
}

// Draw renders the Watches window inside the specified rectangle
func (w *WatchWindow) Draw(screen tcell.Screen, x, y, width, height int, active bool) {
	if !w.Visible || height < 3 {
		return
	}

	frameStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorBorder)
	if !active {
		frameStyle = tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorDarkGray)
	}
	shadowStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)

	// Drop shadow
	for r := y + 1; r < y+height+1; r++ {
		screen.SetContent(x+width, r, ' ', nil, shadowStyle)
		screen.SetContent(x+width+1, r, ' ', nil, shadowStyle)
	}
	for c := x + 2; c < x+width+2; c++ {
		screen.SetContent(c, y+height, ' ', nil, shadowStyle)
	}

	// Corners
	screen.SetContent(x, y, RuneDoubleTopLeft, nil, frameStyle)
	screen.SetContent(x+width-1, y, RuneDoubleTopRight, nil, frameStyle)
	screen.SetContent(x, y+height-1, RuneDoubleBottomLeft, nil, frameStyle)
	screen.SetContent(x+width-1, y+height-1, RuneDoubleBottomRight, nil, frameStyle)

	// Horizontal borders
	for c := x + 1; c < x+width-1; c++ {
		screen.SetContent(c, y, RuneDoubleHorizontal, nil, frameStyle)
		screen.SetContent(c, y+height-1, RuneDoubleHorizontal, nil, frameStyle)
	}

	// Vertical borders
	for r := y + 1; r < y+height-1; r++ {
		screen.SetContent(x, r, RuneDoubleVertical, nil, frameStyle)
		screen.SetContent(x+width-1, r, RuneDoubleVertical, nil, frameStyle)
	}

	// Title: ` 2 Watches `
	title := fmt.Sprintf(" %d Watches ", w.WindowNumber)
	tx := x + (width-len(title))/2
	titleStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorTitle).Bold(true)
	for i, r := range title {
		screen.SetContent(tx+i, y, r, nil, titleStyle)
	}

	// Interior clear
	bgStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorFg)
	varNameStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorYellow).Bold(true)
	varValStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorWhite)
	varTypeStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorLightCyan)
	statusStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorGreen).Bold(true)

	for r := y + 1; r < y+height-1; r++ {
		for c := x + 1; c < x+width-1; c++ {
			screen.SetContent(c, r, ' ', nil, bgStyle)
		}
	}

	// Top line inside watch window: Status text
	if w.StatusText != "" {
		stX := x + 2
		for _, r := range w.StatusText {
			if stX >= x+width-2 {
				break
			}
			screen.SetContent(stX, y+1, r, nil, statusStyle)
			stX += runewidth.RuneWidth(r)
		}
	}

	// List variables
	viewH := height - 3
	for i := 0; i < viewH; i++ {
		varIdx := w.ScrollY + i
		rowY := y + 2 + i
		if varIdx < len(w.Variables) {
			v := w.Variables[varIdx]
			colX := x + 2

			// Name:
			nameStr := v.Name + ": "
			for _, r := range nameStr {
				if colX >= x+width-2 {
					break
				}
				screen.SetContent(colX, rowY, r, nil, varNameStyle)
				colX += runewidth.RuneWidth(r)
			}

			// Value:
			valStr := v.Value
			for _, r := range valStr {
				if colX >= x+width-15 {
					break
				}
				screen.SetContent(colX, rowY, r, nil, varValStyle)
				colX += runewidth.RuneWidth(r)
			}

			// Type (right-ish):
			if v.Type != "" {
				typeStr := " (" + v.Type + ")"
				for _, r := range typeStr {
					if colX >= x+width-2 {
						break
					}
					screen.SetContent(colX, rowY, r, nil, varTypeStyle)
					colX += runewidth.RuneWidth(r)
				}
			}
		} else if len(w.Variables) == 0 && i == 0 {
			hint := "  (No local variables in current scope)"
			for j, r := range hint {
				if x+2+j < x+width-2 {
					screen.SetContent(x+2+j, rowY, r, nil, bgStyle.Foreground(tcell.ColorDarkGray))
				}
			}
		}
	}
}
