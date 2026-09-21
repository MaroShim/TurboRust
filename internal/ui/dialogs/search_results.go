package dialogs

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/MaroShim/TurboRust/internal/compiler"
	"github.com/MaroShim/TurboRust/internal/ui"
)

// SearchResultsDialog displays project search match list and allows jumping
type SearchResultsDialog struct {
	Visible       bool
	Matches       []compiler.SearchMatch
	RootDir       string
	SelectedIndex int
	OnJump        func(match compiler.SearchMatch)
}

func NewSearchResultsDialog() *SearchResultsDialog {
	return &SearchResultsDialog{
		Visible: false,
	}
}

func (d *SearchResultsDialog) Show(matches []compiler.SearchMatch, rootDir string, onJump func(match compiler.SearchMatch)) {
	d.Matches = matches
	d.RootDir = rootDir
	d.SelectedIndex = 0
	d.OnJump = onJump
	d.Visible = true
}

func (d *SearchResultsDialog) Hide() {
	d.Visible = false
}

func (d *SearchResultsDialog) IsVisible() bool {
	return d.Visible
}

func (d *SearchResultsDialog) MoveUp() {
	if d.SelectedIndex > 0 {
		d.SelectedIndex--
	}
}

func (d *SearchResultsDialog) MoveDown() {
	if d.SelectedIndex+1 < len(d.Matches) {
		d.SelectedIndex++
	}
}

func (d *SearchResultsDialog) SelectCurrent() {
	if d.SelectedIndex >= 0 && d.SelectedIndex < len(d.Matches) {
		d.Visible = false
		if d.OnJump != nil {
			d.OnJump(d.Matches[d.SelectedIndex])
		}
	}
}

func (d *SearchResultsDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !d.Visible {
		return
	}
	screen.HideCursor()

	dialogW := screenW - 10
	if dialogW > 76 {
		dialogW = 76
	}
	if dialogW < 50 {
		dialogW = 50
	}
	dialogH := screenH - 8
	if dialogH > 16 {
		dialogH = 16
	}
	if dialogH < 10 {
		dialogH = 10
	}

	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	title := fmt.Sprintf("Search Results (%d matches)", len(d.Matches))
	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, title)

	listY := y + 2
	listH := dialogH - 5
	maxRows := listH

	startIdx := 0
	if d.SelectedIndex >= maxRows {
		startIdx = d.SelectedIndex - maxRows + 1
	}

	for row := 0; row < maxRows; row++ {
		idx := startIdx + row
		currY := listY + row
		if idx >= len(d.Matches) {
			for c := 0; c < dialogW-4; c++ {
				screen.SetContent(x+2+c, currY, ' ', nil, tcell.StyleDefault.Background(ui.ColorDialogBg))
			}
			continue
		}

		m := d.Matches[idx]
		relPath, err := filepath.Rel(d.RootDir, m.File)
		if err != nil {
			relPath = filepath.Base(m.File)
		}

		lineText := fmt.Sprintf("%s:%d:%d: %s", relPath, m.Line, m.Column, m.Snippet)

		itemStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)
		if idx == d.SelectedIndex {
			itemStyle = tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorYellow).Bold(true)
		}

		screenCol := x + 2
		maxScreenCol := x + dialogW - 2

		for c := screenCol; c < maxScreenCol; c++ {
			screen.SetContent(c, currY, ' ', nil, itemStyle)
		}

		for _, r := range lineText {
			rw := runewidth.RuneWidth(r)
			if screenCol+rw > maxScreenCol {
				break
			}
			screen.SetContent(screenCol, currY, r, nil, itemStyle)
			screenCol += rw
		}
	}

	hintStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(tcell.ColorDarkGray)
	hint := "[ Enter: Jump   Esc: Cancel ]"
	hx := x + (dialogW-len(hint))/2
	for i, r := range hint {
		screen.SetContent(hx+i, y+dialogH-2, r, nil, hintStyle)
	}
}

// HandleMouse processes mouse events when SearchResultsDialog is open.
func (d *SearchResultsDialog) HandleMouse(mx, my int, btn tcell.ButtonMask, screenW, screenH int) bool {
	if !d.Visible {
		return false
	}
	dialogW := screenW - 10
	if dialogW > 76 {
		dialogW = 76
	}
	if dialogW < 50 {
		dialogW = 50
	}
	dialogH := screenH - 8
	if dialogH > 16 {
		dialogH = 16
	}
	if dialogH < 10 {
		dialogH = 10
	}

	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	if btn&tcell.WheelUp != 0 {
		d.MoveUp()
		return true
	}
	if btn&tcell.WheelDown != 0 {
		d.MoveDown()
		return true
	}

	if btn&tcell.Button1 != 0 {
		listY := y + 2
		listH := dialogH - 5
		maxRows := listH

		startIdx := 0
		if d.SelectedIndex >= maxRows {
			startIdx = d.SelectedIndex - maxRows + 1
		}

		if my >= listY && my < listY+listH && mx >= x+2 && mx < x+dialogW-2 {
			clickedRow := my - listY
			clickedIdx := startIdx + clickedRow
			if clickedIdx >= 0 && clickedIdx < len(d.Matches) {
				if clickedIdx == d.SelectedIndex {
					d.SelectCurrent()
				} else {
					d.SelectedIndex = clickedIdx
				}
			}
			return true
		}

		if mx >= x && mx < x+dialogW && my >= y && my < y+dialogH {
			return true
		}

		// Click outside: close
		d.Hide()
		return true
	}
	return true
}

