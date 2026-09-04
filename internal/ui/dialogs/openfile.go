package dialogs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"tr/internal/ui"
)

// OpenFileDialog allows selecting or typing a file path to open
type OpenFileDialog struct {
	Visible       bool
	CurrentPath   string
	Files         []string
	SelectedIndex int
	InputText     string
	FocusInput    bool // true: text input, false: file list
	OnOpen        func(path string)
}

func NewOpenFileDialog() *OpenFileDialog {
	return &OpenFileDialog{
		Visible: false,
	}
}

func (o *OpenFileDialog) Show(initialDir string, onOpen func(path string)) {
	if initialDir == "" {
		initialDir = "."
	}
	abs, err := filepath.Abs(initialDir)
	if err == nil {
		o.CurrentPath = abs
	} else {
		o.CurrentPath = initialDir
	}
	o.OnOpen = onOpen
	o.InputText = "*.rs"
	o.FocusInput = false
	o.SelectedIndex = 0
	o.RefreshFiles()
	o.Visible = true
}

func (o *OpenFileDialog) Hide() {
	o.Visible = false
}

func (o *OpenFileDialog) RefreshFiles() {
	entries, err := os.ReadDir(o.CurrentPath)
	o.Files = nil
	if err != nil {
		return
	}

	o.Files = append(o.Files, "..") // Parent directory
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			o.Files = append(o.Files, name+"/")
		} else if strings.HasSuffix(name, ".rs") || strings.HasSuffix(name, ".toml") {
			o.Files = append(o.Files, name)
		}
	}
	if o.SelectedIndex >= len(o.Files) {
		o.SelectedIndex = 0
	}
}

func (o *OpenFileDialog) MoveUp() {
	if o.SelectedIndex > 0 {
		o.SelectedIndex--
	}
}

func (o *OpenFileDialog) MoveDown() {
	if o.SelectedIndex+1 < len(o.Files) {
		o.SelectedIndex++
	}
}

func (o *OpenFileDialog) HandleEnter() {
	if len(o.Files) == 0 {
		return
	}
	selected := o.Files[o.SelectedIndex]
	if selected == ".." {
		o.CurrentPath = filepath.Dir(o.CurrentPath)
		o.RefreshFiles()
		return
	}
	if strings.HasSuffix(selected, "/") {
		o.CurrentPath = filepath.Join(o.CurrentPath, strings.TrimSuffix(selected, "/"))
		o.RefreshFiles()
		return
	}

	fullPath := filepath.Join(o.CurrentPath, selected)
	o.Hide()
	if o.OnOpen != nil {
		o.OnOpen(fullPath)
	}
}

func (o *OpenFileDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !o.Visible {
		return
	}

	dialogW := 50
	dialogH := 16
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "Open a File")

	textStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)
	inputBoxStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	selStyle := tcell.StyleDefault.Background(tcell.ColorTeal).Foreground(tcell.ColorWhite).Bold(true)

	// Current Path display
	pathDisplay := o.CurrentPath
	if len(pathDisplay) > dialogW-6 {
		pathDisplay = "..." + pathDisplay[len(pathDisplay)-(dialogW-9):]
	}
	for i, r := range pathDisplay {
		screen.SetContent(x+3+i, y+2, r, nil, textStyle.Foreground(tcell.ColorNavy))
	}

	// File List Box
	listX := x + 3
	listY := y + 4
	listW := dialogW - 6
	listH := dialogH - 8

	for r := 0; r < listH; r++ {
		fileIdx := r // simple scroll can be added if needed
		curStyle := inputBoxStyle
		lineText := ""
		if fileIdx < len(o.Files) {
			lineText = o.Files[fileIdx]
			if fileIdx == o.SelectedIndex {
				curStyle = selStyle
			}
		}

		for c := 0; c < listW; c++ {
			screen.SetContent(listX+c, listY+r, ' ', nil, curStyle)
		}
		for i, ch := range lineText {
			if i < listW {
				screen.SetContent(listX+i, listY+r, ch, nil, curStyle)
			}
		}
	}

	// Buttons
	ui.DrawButton(screen, x+8, y+dialogH-3, 12, "Open", true)
	ui.DrawButton(screen, x+28, y+dialogH-3, 12, "Cancel", false)
}
