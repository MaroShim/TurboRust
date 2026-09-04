package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// DrawWindowFrame draws a classic Borland double-line window frame with drop shadow
func DrawWindowFrame(
	screen tcell.Screen,
	x, y, width, height int,
	title string,
	windowNum int,
	active bool,
	cursorLine, cursorCol int,
	scrollRatio float64,
) {
	frameStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorBorder)
	if !active {
		frameStyle = tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorDarkGray)
	}
	shadowStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)

	// 1. Draw Drop Shadow (Right side & Bottom side)
	for r := y + 1; r < y+height+1; r++ {
		screen.SetContent(x+width, r, ' ', nil, shadowStyle)
		screen.SetContent(x+width+1, r, ' ', nil, shadowStyle)
	}
	for c := x + 2; c < x+width+2; c++ {
		screen.SetContent(c, y+height, ' ', nil, shadowStyle)
	}

	// 2. Corners
	screen.SetContent(x, y, RuneDoubleTopLeft, nil, frameStyle)
	screen.SetContent(x+width-1, y, RuneDoubleTopRight, nil, frameStyle)
	screen.SetContent(x, y+height-1, RuneDoubleBottomLeft, nil, frameStyle)
	screen.SetContent(x+width-1, y+height-1, RuneDoubleBottomRight, nil, frameStyle)

	// 3. Top Border
	for c := x + 1; c < x+width-1; c++ {
		screen.SetContent(c, y, RuneDoubleHorizontal, nil, frameStyle)
	}

	// 4. Bottom Border
	for c := x + 1; c < x+width-1; c++ {
		screen.SetContent(c, y+height-1, RuneDoubleHorizontal, nil, frameStyle)
	}

	// 5. Left and Right Borders
	for r := y + 1; r < y+height-1; r++ {
		screen.SetContent(x, r, RuneDoubleVertical, nil, frameStyle)
		screen.SetContent(x+width-1, r, RuneDoubleVertical, nil, frameStyle)
	}

	// 6. Close Button [■] at top left: x+2 .. x+4
	if width > 10 {
		btnStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorGreen)
		screen.SetContent(x+2, y, '[', nil, frameStyle)
		screen.SetContent(x+3, y, RuneCloseButton, nil, btnStyle)
		screen.SetContent(x+4, y, ']', nil, frameStyle)
	}

	// 7. Maximize Button [▲] at top right: x+width-5 .. x+width-3
	if width > 16 {
		screen.SetContent(x+width-5, y, '[', nil, frameStyle)
		screen.SetContent(x+width-4, y, RuneZoomButton, nil, frameStyle.Foreground(tcell.ColorGreen))
		screen.SetContent(x+width-3, y, ']', nil, frameStyle)
	}

	// 8. Title: ` 1 FILENAME.RS ` in yellow
	titleStr := fmt.Sprintf(" %d %s ", windowNum, title)
	titleRunes := []rune(titleStr)
	titleLen := len(titleRunes)
	if width > titleLen+14 {
		titleStartX := x + (width-titleLen)/2
		titleStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorTitle).Bold(true)
		if !active {
			titleStyle = tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorLightGray)
		}
		for i, r := range titleRunes {
			screen.SetContent(titleStartX+i, y, r, nil, titleStyle)
		}
	}

	// 9. Bottom status: `[ Line:Col ]`
	posStr := fmt.Sprintf("[ %d:%d ]", cursorLine, cursorCol)
	posRunes := []rune(posStr)
	if width > len(posRunes)+4 {
		posStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorYellow)
		for i, r := range posRunes {
			screen.SetContent(x+2+i, y+height-1, r, nil, posStyle)
		}
	}

	// 10. Scrollbar on right edge
	if height > 4 {
		trackH := height - 4
		thumbPos := int(scrollRatio * float64(trackH-1))
		if thumbPos < 0 {
			thumbPos = 0
		}
		if thumbPos >= trackH {
			thumbPos = trackH - 1
		}

		scrollStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorBorder)
		thumbStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorWhite)

		screen.SetContent(x+width-1, y+1, RuneArrowUp, nil, scrollStyle)
		for r := 0; r < trackH; r++ {
			sy := y + 2 + r
			if r == thumbPos {
				screen.SetContent(x+width-1, sy, RuneScrollThumb, nil, thumbStyle)
			} else {
				screen.SetContent(x+width-1, sy, RuneScrollTrack, nil, scrollStyle)
			}
		}
		screen.SetContent(x+width-1, y+height-2, RuneArrowDown, nil, scrollStyle)
	}
}

// DrawDialogBox draws a classic modal dialog with double line border, light gray bg, and shadow
func DrawDialogBox(screen tcell.Screen, x, y, width, height int, title string) {
	bgStyle := tcell.StyleDefault.Background(ColorDialogBg).Foreground(ColorDialogFg)
	borderStyle := tcell.StyleDefault.Background(ColorDialogBg).Foreground(ColorDialogBorder)
	shadowStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)

	// Drop shadow
	for r := y + 1; r < y+height+1; r++ {
		screen.SetContent(x+width, r, ' ', nil, shadowStyle)
		screen.SetContent(x+width+1, r, ' ', nil, shadowStyle)
	}
	for c := x + 2; c < x+width+2; c++ {
		screen.SetContent(c, y+height, ' ', nil, shadowStyle)
	}

	// Fill background
	for r := y; r < y+height; r++ {
		for c := x; c < x+width; c++ {
			screen.SetContent(c, r, ' ', nil, bgStyle)
		}
	}

	// Double border
	screen.SetContent(x, y, RuneDoubleTopLeft, nil, borderStyle)
	screen.SetContent(x+width-1, y, RuneDoubleTopRight, nil, borderStyle)
	screen.SetContent(x, y+height-1, RuneDoubleBottomLeft, nil, borderStyle)
	screen.SetContent(x+width-1, y+height-1, RuneDoubleBottomRight, nil, borderStyle)

	for c := x + 1; c < x+width-1; c++ {
		screen.SetContent(c, y, RuneDoubleHorizontal, nil, borderStyle)
		screen.SetContent(c, y+height-1, RuneDoubleHorizontal, nil, borderStyle)
	}
	for r := y + 1; r < y+height-1; r++ {
		screen.SetContent(x, r, RuneDoubleVertical, nil, borderStyle)
		screen.SetContent(x+width-1, r, RuneDoubleVertical, nil, borderStyle)
	}

	// Title
	if title != "" {
		t := " " + title + " "
		tx := x + (width-runewidth.StringWidth(t))/2
		titleStyle := tcell.StyleDefault.Background(ColorDialogBg).Foreground(ColorDialogTitle).Bold(true)
		for i, r := range t {
			screen.SetContent(tx+i, y, r, nil, titleStyle)
		}
	}
}

// DrawButton draws a retro dialog button like [ OK ] or [ Cancel ] with shadow
func DrawButton(screen tcell.Screen, x, y, width int, text string, focused bool) {
	btnBg := ColorButtonBg
	btnFg := ColorButtonFg
	if focused {
		btnBg = tcell.ColorWhite
		btnFg = tcell.ColorBlack
	}
	btnStyle := tcell.StyleDefault.Background(btnBg).Foreground(btnFg)
	shadowStyle := tcell.StyleDefault.Background(ColorDialogBg).Foreground(tcell.ColorDarkGray)

	// Button text centered
	btnText := fmt.Sprintf("[ %s ]", text)
	pad := (width - len([]rune(btnText))) / 2

	for c := 0; c < width; c++ {
		screen.SetContent(x+c, y, ' ', nil, btnStyle)
	}
	for i, r := range btnText {
		screen.SetContent(x+pad+i, y, r, nil, btnStyle.Bold(focused))
	}

	// Button shadow (1 char right and 1 row bottom)
	screen.SetContent(x+width, y, '▄', nil, shadowStyle)
	for c := 1; c <= width; c++ {
		screen.SetContent(x+c, y+1, '▀', nil, shadowStyle)
	}
}
