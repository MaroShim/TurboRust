package ui

import (
	"testing"

	"github.com/MaroShim/TurboRust/internal/lsp"
)

func TestCompletionPopupNavigationAndFiltering(t *testing.T) {
	popup := NewCompletionPopup()
	if popup.IsVisible() {
		t.Fatalf("expected popup to start hidden")
	}

	items := []lsp.CompletionItem{
		{Label: "calculate", Kind: lsp.CompletionKindFunction, Detail: "fn(val: i32) -> i32"},
		{Label: "clone", Kind: lsp.CompletionKindMethod, Detail: "fn(&self) -> Self"},
		{Label: "count", Kind: lsp.CompletionKindMethod, Detail: "fn(self) -> usize"},
		{Label: "contains", Kind: lsp.CompletionKindMethod, Detail: "fn(&self, &T) -> bool"},
	}

	popup.Show(items, 4, 10, 5, 80, 24)
	if !popup.IsVisible() {
		t.Fatalf("expected popup to be visible")
	}
	if len(popup.Filtered) != 4 {
		t.Fatalf("expected 4 filtered items, got %d", len(popup.Filtered))
	}

	// 1. Initial selection
	sel := popup.GetSelected()
	if sel == nil || sel.Label != "calculate" {
		t.Fatalf("expected calculate selected, got %+v", sel)
	}

	// 2. Down navigation
	popup.MoveDown()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "clone" {
		t.Fatalf("expected clone selected, got %+v", sel)
	}

	// 3. Up navigation back to 0
	popup.MoveUp()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "calculate" {
		t.Fatalf("expected calculate selected, got %+v", sel)
	}

	// 4. Wrap-around Up navigation (Rule 52)
	popup.MoveUp()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "contains" {
		t.Fatalf("expected contains selected on wrap-around, got %+v", sel)
	}

	// 5. Wrap-around Down navigation
	popup.MoveDown()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "calculate" {
		t.Fatalf("expected calculate selected on wrap-around, got %+v", sel)
	}

	// 6. Filtering with prefix "co"
	popup.SetFilter("co")
	if len(popup.Filtered) != 2 {
		t.Fatalf("expected 2 items matching 'co', got %d", len(popup.Filtered))
	}
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "count" {
		t.Fatalf("expected count selected, got %+v", sel)
	}

	// 7. Filtering with non-matching string hides popup
	popup.SetFilter("zzzz")
	if popup.IsVisible() {
		t.Fatalf("expected popup to hide on zero matches")
	}
}

func TestEditorCompletionHelpers(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{"calc::calc"}
	ed.CursorY = 0
	ed.CursorX = 10 // at end of "calc"

	pref, startCol := ed.GetWordPrefixAtCursor()
	if pref != "calc" || startCol != 6 {
		t.Fatalf("GetWordPrefixAtCursor: got pref=%q, startCol=%d; want pref='calc', startCol=6", pref, startCol)
	}

	ed.ApplyCompletion(startCol, "calculate")
	if ed.Lines[0] != "calc::calculate" {
		t.Fatalf("ApplyCompletion: expected 'calc::calculate', got %q", ed.Lines[0])
	}
	if ed.CursorX != 15 {
		t.Fatalf("ApplyCompletion: expected CursorX=15, got %d", ed.CursorX)
	}

	// Undo rollback verification
	if !ed.Undo() {
		t.Fatalf("Undo failed")
	}
	if ed.Lines[0] != "calc::calc" {
		t.Fatalf("expected rollback to 'calc::calc', got %q", ed.Lines[0])
	}
}
