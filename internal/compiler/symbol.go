package compiler

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchMatch represents a single search match in project
type SearchMatch struct {
	File    string
	Line    int
	Column  int
	Snippet string
}

// GetSearchRootDir determines the root directory to search (cargo root or file directory)
func GetSearchRootDir(targetPath string) string {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}
	if cargoRoot, hasCargo := FindCargoRoot(absTarget); hasCargo {
		return cargoRoot
	}
	fi, err := os.Stat(absTarget)
	if err == nil && !fi.IsDir() {
		return filepath.Dir(absTarget)
	}
	return absTarget
}

// CollectRustFiles gathers all .rs files within rootDir (skipping .git, target, hidden dirs)
func CollectRustFiles(rootDir string) []string {
	var files []string
	_ = filepath.Walk(rootDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			base := fi.Name()
			if strings.HasPrefix(base, ".") || base == "target" || base == "bin" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".rs") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// FindDefinitionInProject searches all .rs files in the project for the definition of symbol
func FindDefinitionInProject(currentFilePath, symbol string) (defFile string, defLine int, defCol int, found bool) {
	if strings.TrimSpace(symbol) == "" {
		return "", 0, 0, false
	}

	rootDir := GetSearchRootDir(currentFilePath)
	files := CollectRustFiles(rootDir)

	// Sort files so the current file is checked first
	if currentFilePath != "" {
		absCurrent, _ := filepath.Abs(currentFilePath)
		for i, f := range files {
			absF, _ := filepath.Abs(f)
			if absF == absCurrent {
				if i > 0 {
					files = append([]string{f}, append(files[:i], files[i+1:]...)...)
				}
				break
			}
		}
	}

	// Rust symbol definition patterns
	escaped := regexp.QuoteMeta(symbol)
	// 1. Functions: fn symbol(...) or fn symbol<...>(
	fnPattern := regexp.MustCompile(`(?m)^\s*(?:pub(?:\([^)]+\))?\s+)?(?:async\s+)?(?:const\s+)?(?:unsafe\s+)?(?:extern(?:\s+"[^"]+")?\s+)?fn\s+` + escaped + `\s*[\(<]`)
	// 2. Struct, enum, trait, type:
	typePattern := regexp.MustCompile(`(?m)^\s*(?:pub(?:\([^)]+\))?\s+)?(?:struct|enum|trait|type|union)\s+` + escaped + `\b`)
	// 3. Const or static:
	constPattern := regexp.MustCompile(`(?m)^\s*(?:pub(?:\([^)]+\))?\s+)?(?:const|static(?:\s+mut)?)\s+` + escaped + `\b`)
	// 4. Macro definition: macro_rules! symbol
	macroPattern := regexp.MustCompile(`(?m)^\s*macro_rules!\s+` + escaped + `\b`)

	patterns := []*regexp.Regexp{fnPattern, typePattern, constPattern, macroPattern}

	for _, p := range patterns {
		for _, f := range files {
			file, err := os.Open(f)
			if err != nil {
				continue
			}
			scanner := bufio.NewScanner(file)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				loc := p.FindStringIndex(line)
				if len(loc) > 0 {
					file.Close()
					col := strings.Index(line, symbol) + 1
					if col <= 0 {
						col = 1
					}
					return f, lineNum, col, true
				}
			}
			file.Close()
		}
	}

	return "", 0, 0, false
}

// SearchInProject searches for query in all .rs files in the project
func SearchInProject(currentFilePath, query string, caseSensitive bool) []SearchMatch {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	rootDir := GetSearchRootDir(currentFilePath)
	files := CollectRustFiles(rootDir)

	var matches []SearchMatch
	targetQuery := query
	if !caseSensitive {
		targetQuery = strings.ToLower(query)
	}

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			searchIn := line
			if !caseSensitive {
				searchIn = strings.ToLower(line)
			}
			idx := strings.Index(searchIn, targetQuery)
			if idx >= 0 {
				col := idx + 1
				snippet := strings.TrimSpace(line)
				if len(snippet) > 60 {
					snippet = snippet[:60] + "..."
				}
				matches = append(matches, SearchMatch{
					File:    f,
					Line:    lineNum,
					Column:  col,
					Snippet: snippet,
				})
			}
		}
		file.Close()
	}

	return matches
}
