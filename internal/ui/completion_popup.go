package ui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/MaroShim/TurboRust/internal/lsp"
)

// CompletionPopup renders an authentic Turbo Vision dropdown autocomplete box.
type CompletionPopup struct {
	Visible       bool
	Items         []lsp.CompletionItem
	Filtered      []lsp.CompletionItem
	SelectedIndex int
	ScrollOffset  int
	TriggerCol    int // 0-based start column in editor line where completion applies
	FilterPrefix  string
	ScreenX       int
	ScreenY       int
	Width         int
	Height        int
}

// NewCompletionPopup creates a new, initially hidden completion popup.
func NewCompletionPopup() *CompletionPopup {
	return &CompletionPopup{
		Visible: false,
		Width:   38,
		Height:  10,
	}
}

// Show opens the completion popup at the requested screen coordinates.
func (cp *CompletionPopup) Show(items []lsp.CompletionItem, triggerCol, screenX, screenY, maxScreenW, maxScreenH int) {
	if len(items) == 0 {
		cp.Hide()
		return
	}

	cp.Items = items
	cp.TriggerCol = triggerCol
	cp.FilterPrefix = ""
	cp.SelectedIndex = 0
	cp.ScrollOffset = 0
	cp.refilter()

	if len(cp.Filtered) == 0 {
		cp.Hide()
		return
	}

	// Calculate optimal box dimensions
	popupW := 38
	for _, it := range cp.Filtered {
		w := runewidth.StringWidth(it.Label) + 12 // padding + badge width
		if w > popupW {
			popupW = w
		}
	}
	if popupW > 55 {
		popupW = 55
	}
	if popupW > maxScreenW-4 {
		popupW = maxScreenW - 4
	}
	if popupW < 24 {
		popupW = 24
	}

	popupH := len(cp.Filtered) + 2 // 2 for borders
	if popupH > 10 {
		popupH = 10
	}

	// Position popup right below cursor; if near bottom, position right above
	posX := screenX
	posY := screenY + 1
	if posX+popupW+2 > maxScreenW {
		posX = maxScreenW - popupW - 2
	}
	if posX < 1 {
		posX = 1
	}

	if posY+popupH+1 >= maxScreenH { // Leave room for status bar (maxScreenH-1)
		posY = screenY - popupH
		if posY < 1 {
			posY = 1
		}
	}

	cp.ScreenX = posX
	cp.ScreenY = posY
	cp.Width = popupW
	cp.Height = popupH
	cp.Visible = true
}

// Hide closes the completion popup.
func (cp *CompletionPopup) Hide() {
	cp.Visible = false
	cp.Items = nil
	cp.Filtered = nil
	cp.SelectedIndex = 0
	cp.ScrollOffset = 0
	cp.FilterPrefix = ""
}

// IsVisible returns whether the popup is currently displayed.
func (cp *CompletionPopup) IsVisible() bool {
	return cp != nil && cp.Visible && len(cp.Filtered) > 0
}

// SetFilter filters the completion candidate list by the typed prefix.
func (cp *CompletionPopup) SetFilter(prefix string) {
	cp.FilterPrefix = prefix
	cp.refilter()
	if len(cp.Filtered) == 0 {
		cp.Hide()
	}
}

func (cp *CompletionPopup) refilter() {
	if cp.FilterPrefix == "" {
		cp.Filtered = make([]lsp.CompletionItem, len(cp.Items))
		copy(cp.Filtered, cp.Items)
	} else {
		lowerPref := strings.ToLower(cp.FilterPrefix)
		var res []lsp.CompletionItem
		for _, it := range cp.Items {
			target := it.FilterText
			if target == "" {
				target = it.Label
			}
			if strings.HasPrefix(strings.ToLower(target), lowerPref) {
				res = append(res, it)
			}
		}
		// If prefix match returned nothing, try substring match
		if len(res) == 0 {
			for _, it := range cp.Items {
				target := it.FilterText
				if target == "" {
					target = it.Label
				}
				if strings.Contains(strings.ToLower(target), lowerPref) {
					res = append(res, it)
				}
			}
		}
		cp.Filtered = res
	}

	if cp.SelectedIndex >= len(cp.Filtered) {
		cp.SelectedIndex = len(cp.Filtered) - 1
	}
	if cp.SelectedIndex < 0 {
		cp.SelectedIndex = 0
	}
	cp.adjustScroll()
}

// MoveDown selects the next completion item (with cyclic wrap-around).
func (cp *CompletionPopup) MoveDown() {
	if len(cp.Filtered) == 0 {
		return
	}
	cp.SelectedIndex++
	if cp.SelectedIndex >= len(cp.Filtered) {
		cp.SelectedIndex = 0 // Wrap around (Rule 52)
	}
	cp.adjustScroll()
}

// MoveUp selects the previous completion item (with cyclic wrap-around).
func (cp *CompletionPopup) MoveUp() {
	if len(cp.Filtered) == 0 {
		return
	}
	cp.SelectedIndex--
	if cp.SelectedIndex < 0 {
		cp.SelectedIndex = len(cp.Filtered) - 1 // Wrap around (Rule 52)
	}
	cp.adjustScroll()
}

// PageDown moves selection down by visible rows.
func (cp *CompletionPopup) PageDown() {
	visibleRows := cp.Height - 2
	if visibleRows < 1 {
		visibleRows = 1
	}
	cp.SelectedIndex += visibleRows
	if cp.SelectedIndex >= len(cp.Filtered) {
		cp.SelectedIndex = len(cp.Filtered) - 1
	}
	cp.adjustScroll()
}

// PageUp moves selection up by visible rows.
func (cp *CompletionPopup) PageUp() {
	visibleRows := cp.Height - 2
	if visibleRows < 1 {
		visibleRows = 1
	}
	cp.SelectedIndex -= visibleRows
	if cp.SelectedIndex < 0 {
		cp.SelectedIndex = 0
	}
	cp.adjustScroll()
}

func (cp *CompletionPopup) adjustScroll() {
	visibleRows := cp.Height - 2
	if visibleRows <= 0 {
		return
	}
	if cp.SelectedIndex < cp.ScrollOffset {
		cp.ScrollOffset = cp.SelectedIndex
	}
	if cp.SelectedIndex >= cp.ScrollOffset+visibleRows {
		cp.ScrollOffset = cp.SelectedIndex - visibleRows + 1
	}
}

// GetSelected returns the currently selected completion item, or nil if none.
func (cp *CompletionPopup) GetSelected() *lsp.CompletionItem {
	if !cp.IsVisible() || cp.SelectedIndex < 0 || cp.SelectedIndex >= len(cp.Filtered) {
		return nil
	}
	return &cp.Filtered[cp.SelectedIndex]
}

// Draw renders the Turbo Vision completion box onto the screen with drop-shadow.
func (cp *CompletionPopup) Draw(screen tcell.Screen) {
	if !cp.IsVisible() {
		return
	}

	x := cp.ScreenX
	y := cp.ScreenY
	w := cp.Width
	h := cp.Height

	bgStyle := tcell.StyleDefault.Background(ColorDialogBg).Foreground(ColorDialogFg)
	borderStyle := tcell.StyleDefault.Background(ColorDialogBg).Foreground(ColorDialogBorder)
	shadowStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)

	// 1. Draw Drop Shadow (Right & Bottom) - Rule 51
	for r := y + 1; r < y+h+1; r++ {
		screen.SetContent(x+w, r, ' ', nil, shadowStyle)
		screen.SetContent(x+w+1, r, ' ', nil, shadowStyle)
	}
	for c := x + 2; c < x+w+2; c++ {
		screen.SetContent(c, y+h, ' ', nil, shadowStyle)
	}

	// 2. Clear background
	for r := y; r < y+h; r++ {
		for c := x; c < x+w; c++ {
			screen.SetContent(c, r, ' ', nil, bgStyle)
		}
	}

	// 3. Draw Double-Line Border
	screen.SetContent(x, y, RuneDoubleTopLeft, nil, borderStyle)
	screen.SetContent(x+w-1, y, RuneDoubleTopRight, nil, borderStyle)
	screen.SetContent(x, y+h-1, RuneDoubleBottomLeft, nil, borderStyle)
	screen.SetContent(x+w-1, y+h-1, RuneDoubleBottomRight, nil, borderStyle)

	for c := x + 1; c < x+w-1; c++ {
		screen.SetContent(c, y, RuneDoubleHorizontal, nil, borderStyle)
		screen.SetContent(c, y+h-1, RuneDoubleHorizontal, nil, borderStyle)
	}
	for r := y + 1; r < y+h-1; r++ {
		screen.SetContent(x, r, RuneDoubleVertical, nil, borderStyle)
		screen.SetContent(x+w-1, r, RuneDoubleVertical, nil, borderStyle)
	}

	// 4. Render Items
	visibleRows := h - 2
	innerW := w - 2
	for i := 0; i < visibleRows; i++ {
		itemIdx := cp.ScrollOffset + i
		rowY := y + 1 + i
		if itemIdx >= len(cp.Filtered) {
			break
		}

		item := cp.Filtered[itemIdx]
		isSelected := (itemIdx == cp.SelectedIndex)

		var rowStyle tcell.Style
		var badgeStyle tcell.Style
		if isSelected {
			// High contrast light cyan selection bar (Rule 52, 56)
			rowStyle = tcell.StyleDefault.Background(tcell.ColorLightCyan).Foreground(tcell.ColorBlack).Bold(true)
			badgeStyle = tcell.StyleDefault.Background(tcell.ColorLightCyan).Foreground(tcell.ColorNavy).Bold(true)
		} else {
			rowStyle = bgStyle
			badgeStyle = tcell.StyleDefault.Background(ColorDialogBg).Foreground(tcell.ColorDarkBlue)
		}

		// Clear row with rowStyle
		for c := x + 1; c < x+w-1; c++ {
			screen.SetContent(c, rowY, ' ', nil, rowStyle)
		}

		// Badge on the right: e.g. [func]
		badge := item.Kind.Badge()
		badgeW := runewidth.StringWidth(badge)
		badgeX := x + w - 1 - badgeW - 1 // 1 space padding from right border
		if badgeX > x+1 {
			for bi, br := range badge {
				screen.SetContent(badgeX+bi, rowY, br, nil, badgeStyle)
			}
		}

		// Item Label on the left
		labelRunes := []rune(item.Label)
		maxLabelW := badgeX - (x + 2) // room between left padding and badge
		if maxLabelW < 5 {
			maxLabelW = innerW - 1
		}

		curX := x + 2
		for _, r := range labelRunes {
			rw := runewidth.RuneWidth(r)
			if (curX - (x + 2) + rw) > maxLabelW {
				break
			}
			screen.SetContent(curX, rowY, r, nil, rowStyle)
			curX += rw
		}
	}

	// 5. Scroll indicators on right border if items exceed visible rows
	if len(cp.Filtered) > visibleRows {
		if cp.ScrollOffset > 0 {
			screen.SetContent(x+w-1, y+1, RuneArrowUp, nil, borderStyle.Foreground(tcell.ColorYellow))
		}
		if cp.ScrollOffset+visibleRows < len(cp.Filtered) {
			screen.SetContent(x+w-1, y+h-2, RuneArrowDown, nil, borderStyle.Foreground(tcell.ColorYellow))
		}
	}
}

// IsWordRune reports whether r is a standard Rust/Go identifier character.
func IsWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// HandleMouse handles mouse events on the completion popup.
// Returns (handled, applied).
func (cp *CompletionPopup) HandleMouse(mx, my int, btn tcell.ButtonMask) (bool, bool) {
	if !cp.IsVisible() {
		return false, false
	}

	if btn&tcell.WheelUp != 0 {
		cp.MoveUp()
		return true, false
	}
	if btn&tcell.WheelDown != 0 {
		cp.MoveDown()
		return true, false
	}

	if btn&tcell.Button1 != 0 {
		// Inside popup bounding box
		if mx >= cp.ScreenX && mx < cp.ScreenX+cp.Width && my >= cp.ScreenY && my < cp.ScreenY+cp.Height {
			// Item rows start at ScreenY + 1 and end before bottom border
			if my >= cp.ScreenY+1 && my < cp.ScreenY+cp.Height-1 {
				clickedRow := my - (cp.ScreenY + 1)
				clickedIdx := cp.ScrollOffset + clickedRow
				if clickedIdx >= 0 && clickedIdx < len(cp.Filtered) {
					if clickedIdx == cp.SelectedIndex {
						return true, true // Apply completion
					}
					cp.SelectedIndex = clickedIdx
					return true, false
				}
			}
			return true, false
		}

		// Clicked outside completion popup: hide popup
		cp.Hide()
		return false, false
	}

	return true, false
}

