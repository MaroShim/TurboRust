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

func TestFindDefinitionRustEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tr_test_edge_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file := filepath.Join(tempDir, "lib.rs")
	code := `// Line 1: Comment
// fn fake_fn() -> i32 { 0 }
pub async fn async_fetch() -> Result<(), ()> {
    Ok(())
}

pub(crate) trait StorageHandler {
    fn save(&self);
}

pub const MAX_BUFFER_SIZE: usize = 4096;
static mut GLOBAL_VAL: i32 = 42;

macro_rules! my_debug_log {
    ($($arg:tt)*) => ()
}
`
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. async fn
	f, line, _, ok := FindDefinitionInProject(file, "async_fetch")
	if !ok || line != 3 {
		t.Errorf("expected async_fetch at line 3, got %s:%d (ok=%v)", f, line, ok)
	}

	// 2. pub(crate) trait
	_, line, _, ok = FindDefinitionInProject(file, "StorageHandler")
	if !ok || line != 7 {
		t.Errorf("expected StorageHandler at line 7, got line %d", line)
	}

	// 3. const & static mut
	_, line, _, ok = FindDefinitionInProject(file, "MAX_BUFFER_SIZE")
	if !ok || line != 11 {
		t.Errorf("expected MAX_BUFFER_SIZE at line 11, got line %d", line)
	}
	_, line, _, ok = FindDefinitionInProject(file, "GLOBAL_VAL")
	if !ok || line != 12 {
		t.Errorf("expected GLOBAL_VAL at line 12, got line %d", line)
	}

	// 4. macro_rules!
	_, line, _, ok = FindDefinitionInProject(file, "my_debug_log")
	if !ok || line != 14 {
		t.Errorf("expected my_debug_log at line 14, got line %d", line)
	}

	// 5. Blank query
	_, _, _, ok = FindDefinitionInProject(file, "   ")
	if ok {
		t.Errorf("expected blank query to return false")
	}
	matches := SearchInProject(file, "", false)
	if len(matches) != 0 {
		t.Errorf("expected empty search matches for empty string")
	}
}
