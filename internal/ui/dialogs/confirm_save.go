package dialogs

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/MaroShim/TurboRust/internal/ui"
)

// ConfirmChoice represents the user's decision in the confirm save dialog
type ConfirmChoice int

const (
	ConfirmYes ConfirmChoice = iota
	ConfirmNo
	ConfirmCancel
)

// ConfirmSaveDialog prompts user when closing or switching away from dirty buffers
type ConfirmSaveDialog struct {
	Visible       bool
	FileName      string
	SelectedIndex int // 0: Yes, 1: No, 2: Cancel
	OnChoice      func(choice ConfirmChoice)
}

func NewConfirmSaveDialog() *ConfirmSaveDialog {
	return &ConfirmSaveDialog{
		Visible:       false,
		SelectedIndex: 0,
	}
}

func (c *ConfirmSaveDialog) Show(fileName string, onChoice func(choice ConfirmChoice)) {
	c.FileName = fileName
	c.SelectedIndex = 0
	c.OnChoice = onChoice
	c.Visible = true
}

func (c *ConfirmSaveDialog) Hide() {
	c.Visible = false
}

func (c *ConfirmSaveDialog) IsVisible() bool {
	return c.Visible
}

func (c *ConfirmSaveDialog) MoveLeft() {
	c.SelectedIndex = (c.SelectedIndex + 2) % 3
}

func (c *ConfirmSaveDialog) MoveRight() {
	c.SelectedIndex = (c.SelectedIndex + 1) % 3
}

func (c *ConfirmSaveDialog) Confirm() {
	c.Visible = false
	if c.OnChoice != nil {
		c.OnChoice(ConfirmChoice(c.SelectedIndex))
	}
}

func (c *ConfirmSaveDialog) Choose(choice ConfirmChoice) {
	c.Visible = false
	if c.OnChoice != nil {
		c.OnChoice(choice)
	}
}

func (c *ConfirmSaveDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !c.Visible {
		return
	}
	screen.HideCursor()

	dialogW := 48
	dialogH := 8
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "Save")

	textStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)

	// Message
	msg := fmt.Sprintf("%s has been modified. Save?", c.FileName)
	if runewidth.StringWidth(msg) > dialogW-4 {
		msg = "Buffer has been modified. Save?"
	}
	msgX := x + (dialogW-runewidth.StringWidth(msg))/2
	for i, r := range msg {
		screen.SetContent(msgX+i, y+2, r, nil, textStyle.Bold(true))
	}

	// Buttons: [ Yes ]  [ No ]  [ Cancel ]
	// Widths: 9, 8, 12.
	btnY := y + 4
	btnX1 := x + 6
	btnX2 := x + 18
	btnX3 := x + 29

	ui.DrawButton(screen, btnX1, btnY, 9, "Yes", c.SelectedIndex == 0)
	ui.DrawButton(screen, btnX2, btnY, 8, "No", c.SelectedIndex == 1)
	ui.DrawButton(screen, btnX3, btnY, 12, "Cancel", c.SelectedIndex == 2)
}

// HandleMouse processes mouse events when ConfirmSaveDialog is open.
func (c *ConfirmSaveDialog) HandleMouse(mx, my int, btn tcell.ButtonMask, screenW, screenH int) bool {
	if !c.Visible {
		return false
	}
	dialogW := 48
	dialogH := 8
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	btnY := y + 4
	btnX1 := x + 6  // Yes, width 9
	btnX2 := x + 18 // No, width 8
	btnX3 := x + 29 // Cancel, width 12

	if btn&tcell.Button1 != 0 {
		if my == btnY {
			if mx >= btnX1 && mx < btnX1+9 {
				c.Choose(ConfirmYes)
				return true
			}
			if mx >= btnX2 && mx < btnX2+8 {
				c.Choose(ConfirmNo)
				return true
			}
			if mx >= btnX3 && mx < btnX3+12 {
				c.Choose(ConfirmCancel)
				return true
			}
		}

		if mx >= x && mx < x+dialogW && my >= y && my < y+dialogH {
			return true
		}

		// Click outside: cancel
		c.Choose(ConfirmCancel)
		return true
	}
	return true
}

