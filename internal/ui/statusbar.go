package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

// StatusItem represents a shortcut key action on the bottom bar
type StatusItem struct {
	KeyName string
	Desc    string
	Action  string
}

// StatusBar renders the classic Borland bottom hotkey strip
type StatusBar struct {
	Items   []StatusItem
	Message string
	MsgTime time.Time
}

func NewStatusBar() *StatusBar {
	return &StatusBar{
		Items: []StatusItem{
			{KeyName: "F1", Desc: "Help", Action: "help_about"},
			{KeyName: "F2", Desc: "Save", Action: "file_save"},
			{KeyName: "F3", Desc: "Open", Action: "file_open"},
			{KeyName: "Alt+F9", Desc: "Compile", Action: "compile_compile"},
			{KeyName: "F9", Desc: "Make", Action: "compile_make"},
			{KeyName: "Ctrl+F9", Desc: "Run", Action: "run_run"},
			{KeyName: "Alt+F5", Desc: "User", Action: "run_userscreen"},
			{KeyName: "F10", Desc: "Menu", Action: "menu_toggle"},
		},
	}
}

func (sb *StatusBar) SetMessage(msg string) {
	sb.Message = msg
	sb.MsgTime = time.Now()
}

// Draw renders the status bar on the bottom row
func (sb *StatusBar) Draw(screen tcell.Screen, y, width int) {
	bgStyle := tcell.StyleDefault.Background(ColorStatusBarBg).Foreground(ColorStatusBarFg)
	keyStyle := tcell.StyleDefault.Background(ColorStatusBarBg).Foreground(ColorStatusBarHotKey).Bold(true)

	// Clear row
	for x := 0; x < width; x++ {
		screen.SetContent(x, y, ' ', nil, bgStyle)
	}

	// If there is an active status message within 3 seconds, display it prominently
	if sb.Message != "" && time.Since(sb.MsgTime) < 3*time.Second {
		tagStyle := tcell.StyleDefault.Background(tcell.ColorNavy).Foreground(tcell.ColorWhite).Bold(true)
		msgStyle := tcell.StyleDefault.Background(ColorStatusBarBg).Foreground(tcell.ColorYellow).Bold(true)
		tag := " [Turbo] "
		xPos := 1
		for _, r := range tag {
			screen.SetContent(xPos, y, r, nil, tagStyle)
			xPos++
		}
		xPos++
		for _, r := range sb.Message {
			if xPos < width-1 {
				screen.SetContent(xPos, y, r, nil, msgStyle)
				xPos++
			}
		}
		return
	}

	xPos := 1
	for _, item := range sb.Items {
		if xPos >= width-2 {
			break
		}

		// Draw KeyName (e.g. F1, Alt+F9)
		for _, r := range item.KeyName {
			if xPos < width {
				screen.SetContent(xPos, y, r, nil, keyStyle)
				xPos++
			}
		}

		// Draw Desc (e.g. Help, Save)
		screen.SetContent(xPos, y, ' ', nil, bgStyle)
		xPos++

		for _, r := range item.Desc {
			if xPos < width {
				screen.SetContent(xPos, y, r, nil, bgStyle)
				xPos++
			}
		}

		// Spacer
		screen.SetContent(xPos, y, ' ', nil, bgStyle)
		xPos++
		screen.SetContent(xPos, y, ' ', nil, bgStyle)
		xPos++
	}
}
