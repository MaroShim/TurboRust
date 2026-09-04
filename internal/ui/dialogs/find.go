package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"tr/internal/ui"
)

// FindDialog represents the retro Borland Find Text dialog
type FindDialog struct {
	Visible       bool
	Query         string
	CaseSensitive bool
	FocusField    int // 0: Query input, 1: CaseSensitive checkbox, 2: OK, 3: Cancel
	OnFind        func(query string, caseSensitive bool)
}

func NewFindDialog() *FindDialog {
	return &FindDialog{
		Visible:       false,
		CaseSensitive: false,
		FocusField:    0,
	}
}

func (f *FindDialog) Show(initialQuery string, onFind func(query string, caseSensitive bool)) {
	if initialQuery != "" {
		f.Query = initialQuery
	}
	f.FocusField = 0
	f.OnFind = onFind
	f.Visible = true
}

func (f *FindDialog) Hide() {
	f.Visible = false
}

func (f *FindDialog) InsertRune(ch rune) {
	if f.FocusField == 0 {
		f.Query += string(ch)
	} else if f.FocusField == 1 && ch == ' ' {
		f.CaseSensitive = !f.CaseSensitive
	}
}

func (f *FindDialog) Backspace() {
	if f.FocusField == 0 {
		runes := []rune(f.Query)
		if len(runes) > 0 {
			f.Query = string(runes[:len(runes)-1])
		}
	}
}

func (f *FindDialog) NextField() {
	f.FocusField = (f.FocusField + 1) % 4
}

func (f *FindDialog) PrevField() {
	f.FocusField = (f.FocusField + 3) % 4
}

func (f *FindDialog) Confirm() {
	if f.FocusField == 3 {
		// Cancel button pressed
		f.Hide()
		return
	}
	if f.Query != "" {
		f.Visible = false
		if f.OnFind != nil {
			f.OnFind(f.Query, f.CaseSensitive)
		}
	} else {
		f.Visible = false
	}
}

func (f *FindDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !f.Visible {
		return
	}

	dialogW := 44
	dialogH := 12
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "Find")

	labelStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)
	inputBoxStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorYellow).Bold(true)
	if f.FocusField != 0 {
		inputBoxStyle = tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	}
	optFocusStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(tcell.ColorYellow).Bold(true)
	hintStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(tcell.ColorDarkGray)

	// Label: "Text to find:"
	label := "Text to find:"
	for i, r := range label {
		screen.SetContent(x+3+i, y+2, r, nil, labelStyle)
	}

	// Input box
	boxW := dialogW - 6
	for c := 0; c < boxW; c++ {
		screen.SetContent(x+3+c, y+3, ' ', nil, inputBoxStyle)
	}
	for i, r := range f.Query {
		if i < boxW {
			screen.SetContent(x+3+i, y+3, r, nil, inputBoxStyle)
		}
	}
	// Cursor marker if input focused
	if f.FocusField == 0 && len(f.Query) < boxW {
		screen.SetContent(x+3+len(f.Query), y+3, '█', nil, inputBoxStyle)
	}

	// Options: "[X] Case sensitive"
	checkChar := ' '
	if f.CaseSensitive {
		checkChar = 'X'
	}
	optText := "[ ] Case sensitive"
	currOptStyle := labelStyle
	if f.FocusField == 1 {
		currOptStyle = optFocusStyle
	}
	for i, r := range optText {
		screen.SetContent(x+3+i, y+5, r, nil, currOptStyle)
	}
	screen.SetContent(x+4, y+5, checkChar, nil, currOptStyle)

	// Navigation Hint
	hint := "[ Tab: Switch field   Enter: OK ]"
	hx := x + (dialogW-len(hint))/2
	for i, r := range hint {
		screen.SetContent(hx+i, y+dialogH-4, r, nil, hintStyle)
	}

	// Buttons
	ui.DrawButton(screen, x+6, y+dialogH-2, 12, "OK", f.FocusField == 0 || f.FocusField == 2)
	ui.DrawButton(screen, x+24, y+dialogH-2, 12, "Cancel", f.FocusField == 3)
}
