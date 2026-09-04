package syntax

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var (
	ColorSyntaxKeyword   = tcell.ColorYellow
	ColorSyntaxType      = tcell.ColorLightCyan
	ColorSyntaxString    = tcell.ColorLightCyan
	ColorSyntaxNumber    = tcell.ColorLightGreen
	ColorSyntaxComment   = tcell.NewHexColor(0x808080) // Gray
	ColorSyntaxMacro     = tcell.ColorGreen
	ColorSyntaxLifetime  = tcell.ColorMaroon
	ColorSyntaxAttribute = tcell.ColorTeal
	ColorSyntaxNormal    = tcell.ColorWhite
)

var rustKeywords = map[string]bool{
	"as": true, "async": true, "await": true, "break": true,
	"const": true, "continue": true, "crate": true, "dyn": true,
	"else": true, "enum": true, "extern": true, "fn": true,
	"for": true, "if": true, "impl": true, "in": true,
	"let": true, "loop": true, "match": true, "mod": true,
	"move": true, "mut": true, "pub": true, "ref": true,
	"return": true, "self": true, "Self": true, "static": true,
	"struct": true, "super": true, "trait": true, "type": true,
	"union": true, "unsafe": true, "use": true, "where": true,
	"while": true, "yield": true,
}

var rustTypes = map[string]bool{
	"i8": true, "i16": true, "i32": true, "i64": true, "i128": true, "isize": true,
	"u8": true, "u16": true, "u32": true, "u64": true, "u128": true, "usize": true,
	"f32": true, "f64": true, "bool": true, "char": true, "str": true,
	"String": true, "Option": true, "Result": true, "Vec": true,
	"Box": true, "Rc": true, "Arc": true, "Cell": true, "RefCell": true,
	"Mutex": true, "RwLock": true,
}

var rustConstants = map[string]bool{
	"true": true, "false": true, "Some": true, "None": true,
	"Ok": true, "Err": true,
}

var rustMacros = map[string]bool{
	"println": true, "eprintln": true, "print": true, "eprint": true,
	"format": true, "panic": true, "assert": true, "assert_eq": true,
	"assert_ne": true, "vec": true, "todo": true, "unimplemented": true,
	"unreachable": true, "cfg": true, "include": true, "include_str": true,
	"include_bytes": true, "env": true, "concat": true, "write": true,
	"writeln": true, "dbg": true,
}

// Token represents a styled span of runes on a line
type Token struct {
	Style tcell.Style
	Char  rune
}

// HighlightLine parses a line of Rust code and returns a slice of styled runes
func HighlightLine(line string, baseStyle tcell.Style, inBlockComment *bool) []Token {
	runes := []rune(line)
	n := len(runes)
	tokens := make([]Token, n)

	keywordStyle := baseStyle.Foreground(ColorSyntaxKeyword).Bold(true)
	typeStyle := baseStyle.Foreground(ColorSyntaxType)
	stringStyle := baseStyle.Foreground(ColorSyntaxString)
	numberStyle := baseStyle.Foreground(ColorSyntaxNumber)
	commentStyle := baseStyle.Foreground(ColorSyntaxComment)
	macroStyle := baseStyle.Foreground(ColorSyntaxMacro).Bold(true)
	constStyle := baseStyle.Foreground(ColorSyntaxNumber).Bold(true)
	lifetimeStyle := baseStyle.Foreground(ColorSyntaxLifetime)
	attrStyle := baseStyle.Foreground(ColorSyntaxAttribute)

	i := 0
	for i < n {
		// Handle block comments across lines
		if inBlockComment != nil && *inBlockComment {
			if i+1 < n && runes[i] == '*' && runes[i+1] == '/' {
				tokens[i] = Token{Style: commentStyle, Char: runes[i]}
				tokens[i+1] = Token{Style: commentStyle, Char: runes[i+1]}
				i += 2
				*inBlockComment = false
				continue
			}
			tokens[i] = Token{Style: commentStyle, Char: runes[i]}
			i++
			continue
		}

		// Single line comment: // ...
		if i+1 < n && runes[i] == '/' && runes[i+1] == '/' {
			for i < n {
				tokens[i] = Token{Style: commentStyle, Char: runes[i]}
				i++
			}
			break
		}

		// Block comment start: /* ... */
		if i+1 < n && runes[i] == '/' && runes[i+1] == '*' {
			tokens[i] = Token{Style: commentStyle, Char: runes[i]}
			tokens[i+1] = Token{Style: commentStyle, Char: runes[i+1]}
			i += 2
			if inBlockComment != nil {
				*inBlockComment = true
			}
			continue
		}

		// Attributes: #[...] or #![...]
		if runes[i] == '#' && i+1 < n && (runes[i+1] == '[' || (runes[i+1] == '!' && i+2 < n && runes[i+2] == '[')) {
			for i < n {
				tokens[i] = Token{Style: attrStyle, Char: runes[i]}
				if runes[i] == ']' {
					i++
					break
				}
				i++
			}
			continue
		}

		// Raw string literal: r"..." or r#"..."#
		if runes[i] == 'r' && i+1 < n && (runes[i+1] == '"' || runes[i+1] == '#') {
			start := i
			hashCount := 0
			idx := i + 1
			for idx < n && runes[idx] == '#' {
				hashCount++
				idx++
			}
			if idx < n && runes[idx] == '"' {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				i = idx + 1
				for i < n {
					if runes[i] == '"' {
						endHashes := 0
						hIdx := i + 1
						for hIdx < n && runes[hIdx] == '#' && endHashes < hashCount {
							endHashes++
							hIdx++
						}
						if endHashes == hashCount {
							for k := start; k <= hIdx; k++ {
								if k < n {
									tokens[k] = Token{Style: stringStyle, Char: runes[k]}
								}
							}
							i = hIdx
							break
						}
					}
					tokens[i] = Token{Style: stringStyle, Char: runes[i]}
					i++
				}
				continue
			}
		}

		// String literal: "..."
		if runes[i] == '"' {
			tokens[i] = Token{Style: stringStyle, Char: runes[i]}
			i++
			escaped := false
			for i < n {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				if escaped {
					escaped = false
				} else if runes[i] == '\\' {
					escaped = true
				} else if runes[i] == '"' {
					i++
					break
				}
				i++
			}
			continue
		}

		// Character or Lifetime: '...' or 'a
		if runes[i] == '\'' {
			// Check if char literal: 'c' or '\n'
			if i+2 < n && runes[i+2] == '\'' {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				tokens[i+1] = Token{Style: stringStyle, Char: runes[i+1]}
				tokens[i+2] = Token{Style: stringStyle, Char: runes[i+2]}
				i += 3
				continue
			} else if i+3 < n && runes[i+1] == '\\' && runes[i+3] == '\'' {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				tokens[i+1] = Token{Style: stringStyle, Char: runes[i+1]}
				tokens[i+2] = Token{Style: stringStyle, Char: runes[i+2]}
				tokens[i+3] = Token{Style: stringStyle, Char: runes[i+3]}
				i += 4
				continue
			} else if i+1 < n && (unicode.IsLetter(runes[i+1]) || runes[i+1] == '_') {
				// Lifetime: 'a, 'static
				tokens[i] = Token{Style: lifetimeStyle, Char: runes[i]}
				i++
				for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
					tokens[i] = Token{Style: lifetimeStyle, Char: runes[i]}
					i++
				}
				continue
			}
		}

		// Number literal: 123, 0x1a, 0b101, 3.14, etc.
		if unicode.IsDigit(runes[i]) || (runes[i] == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			for i < n && (unicode.IsDigit(runes[i]) || unicode.IsLetter(runes[i]) || runes[i] == '.' || runes[i] == '_') {
				tokens[i] = Token{Style: numberStyle, Char: runes[i]}
				i++
			}
			continue
		}

		// Word (Keyword, Type, Macro, Constant, Identifier)
		if unicode.IsLetter(runes[i]) || runes[i] == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			word := string(runes[start:i])

			// Check if followed by '!' (Macro call)
			isMacro := false
			if i < n && runes[i] == '!' {
				if rustMacros[word] || (start > 0 && runes[start-1] != '.') {
					isMacro = true
				}
			}

			var style tcell.Style
			if isMacro {
				style = macroStyle
			} else if rustKeywords[word] {
				style = keywordStyle
			} else if rustTypes[word] {
				style = typeStyle
			} else if rustConstants[word] {
				style = constStyle
			} else if rustMacros[word] {
				style = macroStyle
			} else if len(word) > 0 && unicode.IsUpper(rune(word[0])) {
				// Convention: CamelCase is likely a Type/Struct/Enum
				style = typeStyle
			} else {
				style = baseStyle.Foreground(ColorSyntaxNormal)
			}

			for k := start; k < i; k++ {
				tokens[k] = Token{Style: style, Char: runes[k]}
			}
			if isMacro && i < n && runes[i] == '!' {
				tokens[i] = Token{Style: macroStyle, Char: runes[i]}
				i++
			}
			continue
		}

		// Other punctuation / whitespace
		tokens[i] = Token{Style: baseStyle.Foreground(ColorSyntaxNormal), Char: runes[i]}
		i++
	}

	return tokens
}
