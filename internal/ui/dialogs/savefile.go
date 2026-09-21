package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/MaroShim/TurboRust/internal/ui"
)

// SaveFileDialog allows entering a filename to save
type SaveFileDialog struct {
	Visible  bool
	FileName string
	OnSave   func(path string)
}

func NewSaveFileDialog() *SaveFileDialog {
	return &SaveFileDialog{Visible: false}
}

func (s *SaveFileDialog) Show(defaultName string, onSave func(path string)) {
	s.FileName = defaultName
	s.OnSave = onSave
	s.Visible = true
}

func (s *SaveFileDialog) Hide() {
	s.Visible = false
}

func (s *SaveFileDialog) InsertRune(ch rune) {
	s.FileName += string(ch)
}

func (s *SaveFileDialog) Backspace() {
	runes := []rune(s.FileName)
	if len(runes) > 0 {
		s.FileName = string(runes[:len(runes)-1])
	}
}

func (s *SaveFileDialog) Confirm() {
	if s.FileName != "" {
		s.Visible = false
		if s.OnSave != nil {
			s.OnSave(s.FileName)
		}
	}
}

func (s *SaveFileDialog) IsVisible() bool {
	return s.Visible
}

func (s *SaveFileDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !s.Visible {
		return
	}

	dialogW := 44
	dialogH := 9
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "Save File As")

	textStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)

	// Label
	label := "Save file name:"
	for i, r := range label {
		screen.SetContent(x+3+i, y+2, r, nil, textStyle)
	}

	// Text box
	boxW := dialogW - 6
	ui.DrawInputField(screen, x+3, y+3, boxW, s.FileName, true)

	// Buttons
	ui.DrawButton(screen, x+6, y+5, 12, "OK", true)
	ui.DrawButton(screen, x+24, y+5, 12, "Cancel", false)
}

// HandleMouse processes mouse events when SaveFileDialog is open.
func (s *SaveFileDialog) HandleMouse(mx, my int, btn tcell.ButtonMask, screenW, screenH int) bool {
	if !s.Visible {
		return false
	}
	dialogW := 44
	dialogH := 9
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	if btn&tcell.Button1 != 0 {
		// OK button: x+6, y+5, width 12
		if my == y+5 && mx >= x+6 && mx < x+6+12 {
			s.Confirm()
			return true
		}
		// Cancel button: x+24, y+5, width 12
		if my == y+5 && mx >= x+24 && mx < x+24+12 {
			s.Hide()
			return true
		}

		// Inside dialog
		if mx >= x && mx < x+dialogW && my >= y && my < y+dialogH {
			return true
		}

		// Outside dialog: close
		s.Hide()
		return true
	}
	return true
}

