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

func TestHighlightRustEdgeCases(t *testing.T) {
	baseStyle := tcell.StyleDefault

	// 1. Empty line
	inBlock := false
	tokens := HighlightLine("", baseStyle, &inBlock)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty line, got %d", len(tokens))
	}

	// 2. Block comment spanning lines
	inBlock = false
	line1 := "/* start comment"
	tokens1 := HighlightLine(line1, baseStyle, &inBlock)
	if !inBlock {
		t.Errorf("expected inBlock true after unclosed block comment")
	}
	fg1, _, _ := tokens1[0].Style.Decompose()
	if fg1 != ColorSyntaxComment {
		t.Errorf("expected ColorSyntaxComment for line1, got %v", fg1)
	}

	line2 := "still comment */"
	tokens2 := HighlightLine(line2, baseStyle, &inBlock)
	if inBlock {
		t.Errorf("expected inBlock false after closing */")
	}
	fg2, _, _ := tokens2[len(tokens2)-1].Style.Decompose()
	if fg2 != ColorSyntaxComment {
		t.Errorf("expected ColorSyntaxComment for line2, got %v", fg2)
	}

	// 3. Numbers: decimal, hex, float
	lineNum := "let x = 42 + 0xFF + 3.14;"
	tokensNum := HighlightLine(lineNum, baseStyle, &inBlock)
	numIdx := 8 // '4' in 42
	fgNum, _, _ := tokensNum[numIdx].Style.Decompose()
	if fgNum != ColorSyntaxNumber {
		t.Errorf("expected ColorSyntaxNumber for '42', got %v", fgNum)
	}

	// 4. Lifetime: &'a str
	lineLife := "&'a str"
	tokensLife := HighlightLine(lineLife, baseStyle, &inBlock)
	fgLife, _, _ := tokensLife[1].Style.Decompose() // '\''
	if fgLife != ColorSyntaxLifetime {
		t.Errorf("expected ColorSyntaxLifetime for 'a, got %v", fgLife)
	}

	// 5. Attributes: #[derive(Debug)]
	lineAttr := "#[derive(Debug)]"
	tokensAttr := HighlightLine(lineAttr, baseStyle, &inBlock)
	fgAttr, _, _ := tokensAttr[0].Style.Decompose()
	if fgAttr != ColorSyntaxAttribute {
		t.Errorf("expected ColorSyntaxAttribute for #[derive(Debug)], got %v", fgAttr)
	}

	// 6. Keyword embedded in identifier: 'let_variable' should not be keyword
	lineIdent := "let_variable = 10;"
	tokensIdent := HighlightLine(lineIdent, baseStyle, &inBlock)
	fgIdent, _, _ := tokensIdent[0].Style.Decompose()
	if fgIdent == ColorSyntaxKeyword {
		t.Errorf("identifier 'let_variable' should not be highlighted as keyword")
	}
}
