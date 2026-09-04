package syntax

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHighlightLine(t *testing.T) {
	baseStyle := tcell.StyleDefault

	// Test keyword & function
	line := "fn main() {"
	tokens := HighlightLine(line, baseStyle, nil)
	if len(tokens) != len(line) {
		t.Fatalf("expected length %d, got %d", len(line), len(tokens))
	}
	fg, _, _ := tokens[0].Style.Decompose()
	if fg != ColorSyntaxKeyword {
		t.Errorf("expected keyword color for 'fn', got %v", fg)
	}

	// Test macro
	line2 := `    println!("Hello, Turbo Rust!");`
	tokens2 := HighlightLine(line2, baseStyle, nil)
	// 'p' of println is at index 4
	fg2, _, _ := tokens2[4].Style.Decompose()
	if fg2 != ColorSyntaxMacro {
		t.Errorf("expected macro color for 'println!', got %v", fg2)
	}

	// Test comment
	line3 := "// This is a comment"
	tokens3 := HighlightLine(line3, baseStyle, nil)
	fg3, _, _ := tokens3[0].Style.Decompose()
	if fg3 != ColorSyntaxComment {
		t.Errorf("expected comment color, got %v", fg3)
	}
}
