package ui

import "github.com/gdamore/tcell/v2"

// Box drawing runes for Borland Turbo Vision look
const (
	// Double lines (Windows and Dialogs)
	RuneDoubleTopLeft     = '╔'
	RuneDoubleTopRight    = '╗'
	RuneDoubleBottomLeft  = '╚'
	RuneDoubleBottomRight = '╝'
	RuneDoubleHorizontal  = '═'
	RuneDoubleVertical    = '║'
	RuneDoubleTDown       = '╦'
	RuneDoubleTUp         = '╩'
	RuneDoubleTRight      = '╠'
	RuneDoubleTLeft       = '╣'
	RuneDoubleCross       = '╬'

	// Single lines (Dropdown menus, separators)
	RuneSingleTopLeft     = '┌'
	RuneSingleTopRight    = '┐'
	RuneSingleBottomLeft  = '└'
	RuneSingleBottomRight = '┘'
	RuneSingleHorizontal  = '─'
	RuneSingleVertical    = '│'
	RuneSingleTRight      = '├'
	RuneSingleTLeft       = '┤'

	// Turbo Vision UI Symbols
	RuneCloseButton = '■' // Window close [■]
	RuneZoomButton  = '▲' // Window maximize [▲]
	RuneArrowDown   = '▼'
	RuneArrowUp     = '▲'
	RuneArrowRight  = '►'
	RuneScrollTrack = '░'
	RuneScrollThumb = '█'
	RuneBreakpoint  = '●'
	RuneCurrentLine = '►'
	RuneShadow      = ' ' // Shadow character with black background
)

// Turbo Vision Color Palette
var (
	// Desktop Background
	ColorDesktopBg = tcell.NewHexColor(0x008080) // Classic Teal or Dark Cyan Pattern
	ColorDesktopFg = tcell.NewHexColor(0x004040)

	// Menu Bar (Top)
	ColorMenuBarBg      = tcell.ColorLightGray
	ColorMenuBarFg      = tcell.ColorBlack
	ColorMenuBarHotKey  = tcell.ColorMaroon
	ColorMenuSelectedBg = tcell.ColorTeal
	ColorMenuSelectedFg = tcell.ColorWhite

	// Dropdown Menu
	ColorDropdownBg       = tcell.ColorLightGray
	ColorDropdownFg       = tcell.ColorBlack
	ColorDropdownBorder   = tcell.ColorBlack
	ColorDropdownHotKey   = tcell.ColorMaroon
	ColorDropdownSelectBg = tcell.ColorTeal
	ColorDropdownSelectFg = tcell.ColorWhite
	ColorDropdownDisabled = tcell.ColorDarkGray

	// Editor Window (Classic Turbo Pascal / Turbo C Blue)
	ColorEditorBg         = tcell.NewHexColor(0x0000A8) // Classic Borland Blue
	ColorEditorFg         = tcell.ColorWhite
	ColorEditorBorder     = tcell.ColorLightCyan
	ColorEditorTitle      = tcell.ColorYellow
	ColorEditorLineNumBg  = tcell.NewHexColor(0x000080)
	ColorEditorLineNumFg  = tcell.ColorLightCyan
	ColorEditorBreakpoint = tcell.ColorRed
	ColorEditorCurrentIP  = tcell.ColorYellow

	// Status Bar (Bottom)
	ColorStatusBarBg     = tcell.ColorLightGray
	ColorStatusBarFg     = tcell.ColorBlack
	ColorStatusBarHotKey = tcell.ColorMaroon

	// Dialog (Modal)
	ColorDialogBg     = tcell.ColorLightGray
	ColorDialogFg     = tcell.ColorBlack
	ColorDialogBorder = tcell.ColorBlack
	ColorDialogTitle  = tcell.ColorBlack
	ColorButtonBg     = tcell.ColorTeal
	ColorButtonFg     = tcell.ColorWhite
	ColorButtonHotKey = tcell.ColorYellow

	// Syntax Highlighting (VS Code Dark+ style with Borland high-contrast harmony)
	ColorSyntaxKeyword  = tcell.NewHexColor(0xFF79C6) // Pink / Magenta (VS Code #C586C0)
	ColorSyntaxFunction = tcell.ColorYellow          // Bright Yellow (VS Code #DCDCAA)
	ColorSyntaxType     = tcell.ColorLightCyan       // Mint / Cyan (VS Code #4EC9B0)
	ColorSyntaxString   = tcell.NewHexColor(0xFFB86C) // Warm Peach / Orange (VS Code #CE9178)
	ColorSyntaxNumber   = tcell.ColorLightGreen      // Soft Green (VS Code #B5CEA8)
	ColorSyntaxComment  = tcell.NewHexColor(0x7EC684) // Calming Green (VS Code #6A9955)
	ColorSyntaxMacro    = tcell.ColorYellow          // Yellow
	ColorSyntaxVariable = tcell.NewHexColor(0x9CDCFE) // Sky Blue (VS Code #9CDCFE)
	ColorSyntaxNormal   = tcell.ColorWhite

	// User Screen
	ColorUserScreenBg = tcell.ColorBlack
	ColorUserScreenFg = tcell.ColorLightGray
)
