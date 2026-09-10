package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDefinitionInRustProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tr_test_def_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "main.rs")
	file2 := filepath.Join(tempDir, "calc.rs")

	code1 := `fn main() {
    println!("{}", add(1, 2));
}
`
	code2 := `// Add two numbers
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}

pub struct Point {
    x: i32,
    y: i32,
}
`
	if err := os.WriteFile(file1, []byte(code1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(code2), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Find add definition from main.rs
	f, line, col, ok := FindDefinitionInProject(file1, "add")
	if !ok {
		t.Fatalf("expected to find definition for add")
	}
	if filepath.Base(f) != "calc.rs" {
		t.Errorf("expected calc.rs, got %s", f)
	}
	if line != 2 {
		t.Errorf("expected line 2, got %d", line)
	}
	if col < 1 {
		t.Errorf("invalid col %d", col)
	}

	// Test 2: Find Point definition
	f, line, _, ok = FindDefinitionInProject(file1, "Point")
	if !ok {
		t.Fatalf("expected to find definition for Point")
	}
	if line != 6 {
		t.Errorf("expected line 6, got %d", line)
	}

	// Test 3: Search in project
	matches := SearchInProject(file1, "add", false)
	if len(matches) < 2 {
		t.Errorf("expected at least 2 matches for add, got %d", len(matches))
	}
}
