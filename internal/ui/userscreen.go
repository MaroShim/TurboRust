package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// UserScreen represents the simulated MS-DOS console screen (Alt+F5 in Turbo C)
type UserScreen struct {
	Active   bool
	Output   string
	ExitCode int
	Duration string
	ScrollY  int
	Lines    []string
}

func NewUserScreen() *UserScreen {
	return &UserScreen{
		Active: false,
		Output: "",
		Lines:  []string{"No program has been executed yet."},
	}
}

// SetLiveOutput updates user screen while program is actively running or paused in debugger
func (u *UserScreen) SetLiveOutput(output string, currentFile string, currentLine int) {
	u.Output = output
	rawLines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	u.Lines = make([]string, 0, len(rawLines)+3)
	if len(rawLines) == 0 || (len(rawLines) == 1 && rawLines[0] == "") {
		u.Lines = append(u.Lines, "[No program output produced yet]")
	} else {
		u.Lines = append(u.Lines, rawLines...)
	}
	u.Lines = append(u.Lines, "")
	if currentFile != "" && currentLine > 0 {
		u.Lines = append(u.Lines, fmt.Sprintf("--- Debugger: paused at %s:%d (Alt+F5 or any key to return to IDE) ---", filepath.Base(currentFile), currentLine))
	} else {
		u.Lines = append(u.Lines, "--- Debugger: active (Alt+F5 or any key to return to IDE) ---")
	}
	u.ScrollY = 0
}

// SetExecutionResult sets the console output from a completed run
func (u *UserScreen) SetExecutionResult(output string, exitCode int, duration string) {
	u.Output = output
	u.ExitCode = exitCode
	u.Duration = duration

	rawLines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	u.Lines = make([]string, 0, len(rawLines)+3)
	if len(rawLines) == 0 || (len(rawLines) == 1 && rawLines[0] == "") {
		u.Lines = append(u.Lines, "[Program produced no output]")
	} else {
		u.Lines = append(u.Lines, rawLines...)
	}
	u.Lines = append(u.Lines, "")
	u.Lines = append(u.Lines, fmt.Sprintf("Process exited with code %d (elapsed: %s)", exitCode, duration))
	u.ScrollY = 0
}

// Show opens the user screen
func (u *UserScreen) Show() {
	u.Active = true
}

// Hide closes the user screen and returns to IDE
func (u *UserScreen) Hide() {
	u.Active = false
}

// ScrollUp moves output view up
func (u *UserScreen) ScrollUp() {
	if u.ScrollY > 0 {
		u.ScrollY--
	}
}

// ScrollDown moves output view down
func (u *UserScreen) ScrollDown(height int) {
	if u.ScrollY+height < len(u.Lines) {
		u.ScrollY++
	}
}

// Draw renders the user screen over the terminal
func (u *UserScreen) Draw(screen tcell.Screen, width, height int) {
	bgStyle := tcell.StyleDefault.Background(ColorUserScreenBg).Foreground(ColorUserScreenFg)
	barStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorYellow).Bold(true)

	// Clear full screen with black
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			screen.SetContent(x, y, ' ', nil, bgStyle)
		}
	}

	// Draw lines
	viewH := height - 1
	for r := 0; r < viewH; r++ {
		lineIdx := u.ScrollY + r
		if lineIdx < len(u.Lines) {
			lineStr := u.Lines[lineIdx]
			x := 0
			for _, ch := range lineStr {
				if x >= width {
					break
				}
				w := runewidth.RuneWidth(ch)
				screen.SetContent(x, r, ch, nil, bgStyle)
				x += w
			}
		}
	}

	// Bottom status banner
	banner := " [ Turbo Rust User Screen (Alt+F5) - Press any key to return to IDE ] "
	bx := (width - len(banner)) / 2
	if bx < 0 {
		bx = 0
	}
	for i, r := range banner {
		if bx+i < width {
			screen.SetContent(bx+i, height-1, r, nil, barStyle)
		}
	}
}
