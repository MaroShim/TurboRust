package ui

import (
	"github.com/gdamore/tcell/v2"
)

// MenuItem represents a single selectable item inside a dropdown menu
type MenuItem struct {
	Label    string
	Shortcut string
	ActionID string
	Disabled bool
	IsSep    bool
}

// Menu represents a top-level menu column
type Menu struct {
	Title    string
	HotKey   rune
	HotIndex int
	Items    []MenuItem
}

// MenuBar manages the top-level menu system
type MenuBar struct {
	Menus        []Menu
	Active       bool
	OpenDropdown bool
	ActiveMenu   int
	ActiveItem   int
}

func NewMenuBar() *MenuBar {
	return &MenuBar{
		Menus: []Menu{
			{
				Title:    "File",
				HotKey:   'F',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "New", ActionID: "file_new"},
					{Label: "Open...", Shortcut: "F3", ActionID: "file_open"},
					{Label: "Save", Shortcut: "F2", ActionID: "file_save"},
					{Label: "Save as...", ActionID: "file_save_as"},
					{IsSep: true},
					{Label: "Exit", Shortcut: "Alt+X", ActionID: "app_exit"},
				},
			},
			{
				Title:    "Edit",
				HotKey:   'E',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Undo", Shortcut: "Alt+BkSp", ActionID: "edit_undo"},
					{Label: "Cut", Shortcut: "Shift+Del", ActionID: "edit_cut"},
					{Label: "Copy", Shortcut: "Ctrl+Ins", ActionID: "edit_copy"},
					{Label: "Paste", Shortcut: "Shift+Ins", ActionID: "edit_paste"},
					{Label: "Clear", Shortcut: "Ctrl+Del", ActionID: "edit_clear"},
				},
			},
			{
				Title:    "Search",
				HotKey:   'S',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Find...", Shortcut: "Ctrl+F", ActionID: "search_find"},
					{Label: "Search again", Shortcut: "Ctrl+L", ActionID: "search_again"},
					{Label: "Replace...", ActionID: "search_replace"},
					{Label: "Go to line...", Shortcut: "Alt+G", ActionID: "search_goto"},
				},
			},
			{
				Title:    "Run",
				HotKey:   'R',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Run", Shortcut: "Ctrl+F9", ActionID: "run_run"},
					{Label: "Program reset", Shortcut: "Ctrl+F2", ActionID: "run_reset"},
					{Label: "User screen", Shortcut: "Alt+F5", ActionID: "run_userscreen"},
				},
			},
			{
				Title:    "Compile",
				HotKey:   'C',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Compile", Shortcut: "Alt+F9", ActionID: "compile_compile"},
					{Label: "Make", Shortcut: "F9", ActionID: "compile_make"},
					{Label: "Build all", ActionID: "compile_buildall"},
				},
			},
			{
				Title:    "Debug",
				HotKey:   'D',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Start / Continue", Shortcut: "F5", ActionID: "debug_continue"},
					{Label: "Step Over", Shortcut: "F8", ActionID: "debug_step_over"},
					{Label: "Trace Into", Shortcut: "F7", ActionID: "debug_step_into"},
					{Label: "Toggle Breakpoint", Shortcut: "F4", ActionID: "debug_toggle_bp"},
					{Label: "Stop Debugger", Shortcut: "Ctrl+F2", ActionID: "debug_stop"},
					{Label: "Watches Window", ActionID: "debug_watches"},
				},
			},
			{
				Title:    "Options",
				HotKey:   'O',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Line Numbers", Shortcut: "Alt+L", ActionID: "options_toggle_linenums"},
					{Label: "Sound: ON", ActionID: "options_toggle_sound"},
					{Label: "Tab Size: 4", ActionID: "options_tab_size"},
				},
			},
			{
				Title:    "Window",
				HotKey:   'W',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Tile", ActionID: "window_tile"},
					{Label: "Cascade", ActionID: "window_cascade"},
					{Label: "Close", Shortcut: "Alt+F3", ActionID: "window_close"},
				},
			},
			{
				Title:    "Help",
				HotKey:   'H',
				HotIndex: 0,
				Items: []MenuItem{
					{Label: "Help Index", Shortcut: "Shift+F1", ActionID: "help_about"},
					{Label: "About Turbo Rust...", ActionID: "help_about"},
				},
			},
		},
		Active:       false,
		OpenDropdown: false,
		ActiveMenu:   0,
		ActiveItem:   0,
	}
}

// Open activates the menubar on F10
func (m *MenuBar) Open() {
	m.OpenMenu(0)
}

// OpenMenu activates the menubar and opens the menu at index
func (m *MenuBar) OpenMenu(index int) {
	if index >= 0 && index < len(m.Menus) {
		m.ActiveMenu = index
	}
	m.Active = true
	m.OpenDropdown = true
	m.ActiveItem = 0
}

// Close closes the menu
func (m *MenuBar) Close() {
	m.Active = false
	m.OpenDropdown = false
}

// MoveLeft moves to previous menu
func (m *MenuBar) MoveLeft() {
	m.ActiveMenu--
	if m.ActiveMenu < 0 {
		m.ActiveMenu = len(m.Menus) - 1
	}
	m.ActiveItem = 0
}

// MoveRight moves to next menu
func (m *MenuBar) MoveRight() {
	m.ActiveMenu++
	if m.ActiveMenu >= len(m.Menus) {
		m.ActiveMenu = 0
	}
	m.ActiveItem = 0
}

// MoveUp moves up within dropdown
func (m *MenuBar) MoveUp() {
	if !m.OpenDropdown {
		return
	}
	items := m.Menus[m.ActiveMenu].Items
	for i := 0; i < len(items); i++ {
		m.ActiveItem--
		if m.ActiveItem < 0 {
			m.ActiveItem = len(items) - 1
		}
		if !items[m.ActiveItem].IsSep {
			break
		}
	}
}

// MoveDown moves down within dropdown
func (m *MenuBar) MoveDown() {
	if !m.OpenDropdown {
		m.OpenDropdown = true
		m.ActiveItem = 0
		return
	}
	items := m.Menus[m.ActiveMenu].Items
	for i := 0; i < len(items); i++ {
		m.ActiveItem++
		if m.ActiveItem >= len(items) {
			m.ActiveItem = 0
		}
		if !items[m.ActiveItem].IsSep {
			break
		}
	}
}

// GetSelectedAction returns the action ID of the currently selected dropdown item
func (m *MenuBar) GetSelectedAction() string {
	if !m.OpenDropdown {
		return ""
	}
	menu := m.Menus[m.ActiveMenu]
	if m.ActiveItem >= 0 && m.ActiveItem < len(menu.Items) {
		item := menu.Items[m.ActiveItem]
		if !item.IsSep && !item.Disabled {
			return item.ActionID
		}
	}
	return ""
}

// Draw renders the top menubar and active dropdown
func (m *MenuBar) Draw(screen tcell.Screen, width int) {
	barStyle := tcell.StyleDefault.Background(ColorMenuBarBg).Foreground(ColorMenuBarFg)
	hotStyle := tcell.StyleDefault.Background(ColorMenuBarBg).Foreground(ColorMenuBarHotKey).Bold(true)
	activeStyle := tcell.StyleDefault.Background(ColorMenuSelectedBg).Foreground(ColorMenuSelectedFg)

	// 1. Clear menu bar row
	for x := 0; x < width; x++ {
		screen.SetContent(x, 0, ' ', nil, barStyle)
	}

	// 2. Draw Top Menus
	xPos := 1
	menuPositions := make([]int, len(m.Menus))

	for i, menu := range m.Menus {
		menuPositions[i] = xPos
		title := menu.Title
		isSelected := m.Active && (m.ActiveMenu == i)

		style := barStyle
		currHotStyle := hotStyle
		if isSelected {
			style = activeStyle
			currHotStyle = activeStyle.Foreground(tcell.ColorYellow).Bold(true)
		}

		screen.SetContent(xPos, 0, ' ', nil, style)
		xPos++

		for j, r := range title {
			if j == menu.HotIndex {
				screen.SetContent(xPos, 0, r, nil, currHotStyle)
			} else {
				screen.SetContent(xPos, 0, r, nil, style)
			}
			xPos++
		}

		screen.SetContent(xPos, 0, ' ', nil, style)
		xPos++
		xPos++
	}

	// 3. Draw Dropdown if open
	if m.Active && m.OpenDropdown {
		m.drawDropdown(screen, menuPositions[m.ActiveMenu])
	}
}

func (m *MenuBar) drawDropdown(screen tcell.Screen, startX int) {
	menu := m.Menus[m.ActiveMenu]

	maxLabel := 0
	maxShort := 0
	for _, item := range menu.Items {
		if item.IsSep {
			continue
		}
		if len(item.Label) > maxLabel {
			maxLabel = len(item.Label)
		}
		if len(item.Shortcut) > maxShort {
			maxShort = len(item.Shortcut)
		}
	}

	contentWidth := maxLabel + maxShort + 4
	if contentWidth < 18 {
		contentWidth = 18
	}
	ddWidth := contentWidth + 2
	ddHeight := len(menu.Items) + 2

	bgStyle := tcell.StyleDefault.Background(ColorDropdownBg).Foreground(ColorDropdownFg)
	borderStyle := tcell.StyleDefault.Background(ColorDropdownBg).Foreground(ColorDropdownBorder)
	selectStyle := tcell.StyleDefault.Background(ColorDropdownSelectBg).Foreground(ColorDropdownSelectFg)
	shadowStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)

	// Draw Drop shadow
	for r := 2; r < ddHeight+2; r++ {
		screen.SetContent(startX+ddWidth, r, ' ', nil, shadowStyle)
		screen.SetContent(startX+ddWidth+1, r, ' ', nil, shadowStyle)
	}
	for c := startX + 2; c < startX+ddWidth+2; c++ {
		screen.SetContent(c, ddHeight+1, ' ', nil, shadowStyle)
	}

	// Clear Box
	for r := 1; r <= ddHeight; r++ {
		for c := startX; c < startX+ddWidth; c++ {
			screen.SetContent(c, r, ' ', nil, bgStyle)
		}
	}

	// Single Line Border
	screen.SetContent(startX, 1, RuneSingleTopLeft, nil, borderStyle)
	screen.SetContent(startX+ddWidth-1, 1, RuneSingleTopRight, nil, borderStyle)
	screen.SetContent(startX, ddHeight, RuneSingleBottomLeft, nil, borderStyle)
	screen.SetContent(startX+ddWidth-1, ddHeight, RuneSingleBottomRight, nil, borderStyle)

	for c := startX + 1; c < startX+ddWidth-1; c++ {
		screen.SetContent(c, 1, RuneSingleHorizontal, nil, borderStyle)
		screen.SetContent(c, ddHeight, RuneSingleHorizontal, nil, borderStyle)
	}
	for r := 2; r < ddHeight; r++ {
		screen.SetContent(startX, r, RuneSingleVertical, nil, borderStyle)
		screen.SetContent(startX+ddWidth-1, r, RuneSingleVertical, nil, borderStyle)
	}

	// Draw Items
	for idx, item := range menu.Items {
		row := 2 + idx
		if item.IsSep {
			screen.SetContent(startX, row, RuneSingleTRight, nil, borderStyle)
			screen.SetContent(startX+ddWidth-1, row, RuneSingleTLeft, nil, borderStyle)
			for c := startX + 1; c < startX+ddWidth-1; c++ {
				screen.SetContent(c, row, RuneSingleHorizontal, nil, borderStyle)
			}
			continue
		}

		isItemActive := (idx == m.ActiveItem)
		itemStyle := bgStyle
		if isItemActive {
			itemStyle = selectStyle
		}

		for c := startX + 1; c < startX+ddWidth-1; c++ {
			screen.SetContent(c, row, ' ', nil, itemStyle)
		}

		labelRunes := []rune(item.Label)
		lx := startX + 2
		for j, r := range labelRunes {
			if !isItemActive && j == 0 && item.ActionID != "" {
				screen.SetContent(lx+j, row, r, nil, itemStyle.Foreground(ColorDropdownHotKey).Bold(true))
			} else {
				screen.SetContent(lx+j, row, r, nil, itemStyle)
			}
		}

		if item.Shortcut != "" {
			sx := startX + ddWidth - 2 - len(item.Shortcut)
			for j, r := range item.Shortcut {
				screen.SetContent(sx+j, row, r, nil, itemStyle)
			}
		}
	}
}

// SetSoundEnabled updates the Sound item label in Options menu
func (m *MenuBar) SetSoundEnabled(enabled bool) {
	for i := range m.Menus {
		if m.Menus[i].Title == "Options" {
			for j := range m.Menus[i].Items {
				if m.Menus[i].Items[j].ActionID == "options_toggle_sound" {
					if enabled {
						m.Menus[i].Items[j].Label = "Sound: ON"
					} else {
						m.Menus[i].Items[j].Label = "Sound: OFF"
					}
					return
				}
			}
		}
	}
}
