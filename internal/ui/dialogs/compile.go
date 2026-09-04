package dialogs

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"tr/internal/compiler"
	"tr/internal/ui"
)

// CompileDialog represents the classic Borland Compiling status box
type CompileDialog struct {
	Visible   bool
	FileName  string
	Lines     int
	Result    *compiler.BuildResult
	Finished  bool
	OnDismiss func()
}

func NewCompileDialog() *CompileDialog {
	return &CompileDialog{
		Visible: false,
	}
}

func (cd *CompileDialog) Show(fileName string, lines int, res *compiler.BuildResult) {
	cd.Visible = true
	cd.FileName = fileName
	cd.Lines = lines
	cd.Result = res
	cd.Finished = true
}

func (cd *CompileDialog) Hide() {
	cd.Visible = false
	if cd.OnDismiss != nil {
		cd.OnDismiss()
	}
}

// Draw renders the modal Compiling dialog centered on screen
func (cd *CompileDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !cd.Visible {
		return
	}

	dialogW := 44
	dialogH := 13
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "Compiling")

	textStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)
	labelStyle := textStyle.Bold(true)

	// Content lines
	drawRow := func(row int, label, val string, valStyle tcell.Style) {
		ry := y + 2 + row
		// Draw label
		lx := x + 3
		for i, r := range label {
			screen.SetContent(lx+i, ry, r, nil, labelStyle)
		}
		// Draw value
		vx := x + 18
		for i, r := range val {
			screen.SetContent(vx+i, ry, r, nil, valStyle)
		}
	}

	drawRow(0, "Main file:", cd.FileName, textStyle)
	drawRow(1, "Compiling:", "rustc -> binary", textStyle)
	drawRow(2, "Total lines:", fmt.Sprintf("%d", cd.Lines), textStyle)

	errCount := 0
	warnCount := 0
	if cd.Result != nil {
		errCount = cd.Result.ErrorCount
		warnCount = cd.Result.WarningCount
	}

	errColor := tcell.ColorGreen
	if errCount > 0 {
		errColor = tcell.ColorRed
	}
	drawRow(3, "Total errors:", fmt.Sprintf("%d", errCount), textStyle.Foreground(errColor).Bold(true))
	drawRow(4, "Warnings:", fmt.Sprintf("%d", warnCount), textStyle)

	elapsed := "0.00s"
	statusStr := "Compiling..."
	statusStyle := textStyle.Foreground(tcell.ColorYellow).Bold(true)

	if cd.Result != nil {
		elapsed = fmt.Sprintf("%.2fs", cd.Result.Duration.Seconds())
		if cd.Result.Success {
			statusStr = "Success"
			statusStyle = textStyle.Foreground(tcell.ColorGreen).Bold(true)
		} else {
			statusStr = "Failed (Errors)"
			statusStyle = textStyle.Foreground(tcell.ColorRed).Bold(true)
		}
	}

	drawRow(5, "Elapsed time:", elapsed, textStyle)
	drawRow(6, "Status:", statusStr, statusStyle)

	// Bottom action prompt
	prompt := "[ Press Any Key to Continue ]"
	if errCount > 0 {
		prompt = "[ Press Enter for Error List ]"
	}
	px := x + (dialogW-len(prompt))/2
	promptStyle := textStyle.Foreground(tcell.ColorNavy).Bold(true)
	for i, r := range prompt {
		screen.SetContent(px+i, y+dialogH-2, r, nil, promptStyle)
	}
}
