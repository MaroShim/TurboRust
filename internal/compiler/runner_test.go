package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseErrorsShort(t *testing.T) {
	output := `src/main.rs:10:5: error[E0425]: cannot find value 'x' in this scope
src/main.rs:15:9: warning: unused variable: 'y'`

	errs := ParseErrors(output)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
	if errs[0].Line != 10 || errs[0].Column != 5 || errs[0].Level != "error" {
		t.Errorf("unexpected error 0: %+v", errs[0])
	}
	if errs[1].Line != 15 || errs[1].Column != 9 || errs[1].Level != "warning" {
		t.Errorf("unexpected warning 1: %+v", errs[1])
	}
}

func TestParseErrorsStandard(t *testing.T) {
	output := `error[E0425]: cannot find value 'x' in this scope
  --> src/main.rs:12:7
   |
12 |     x + 1
   |     ^ not found in this scope`

	errs := ParseErrors(output)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Line != 12 || errs[0].Column != 7 {
		t.Errorf("unexpected line/col: %+v", errs[0])
	}
}

func TestRealRustcBuildAndRun(t *testing.T) {
	helloPath := "../../examples/hello.rs"
	res := Build(helloPath)
	if !res.Success {
		t.Fatalf("expected successful rustc build of %s, got raw: %s", helloPath, res.RawOutput)
	}
	if res.LinesCompiled <= 0 {
		t.Errorf("expected >0 lines compiled, got %d", res.LinesCompiled)
	}

	runRes := Run(res.BinaryPath)
	if !runRes.Completed || runRes.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", runRes.ExitCode)
	}
	if len(runRes.Output) == 0 {
		t.Errorf("expected non-empty output")
	}
}

func TestCargoProjectBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create Cargo.toml
	cargoToml := `[package]
name = "my_app"
version = "0.1.0"
edition = "2021"

[dependencies]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644); err != nil {
		t.Fatalf("failed to write Cargo.toml: %v", err)
	}

	// 2. Create src/main.rs and src/sub.rs
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to mkdir src: %v", err)
	}

	mainCode := `mod sub;
fn main() {
    println!("{}", sub::greet());
}
`
	if err := os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte(mainCode), 0644); err != nil {
		t.Fatalf("failed to write main.rs: %v", err)
	}

	subCode := `pub fn greet() -> &'static str {
    "TURBO RUST CARGO OK"
}
`
	if err := os.WriteFile(filepath.Join(srcDir, "sub.rs"), []byte(subCode), 0644); err != nil {
		t.Fatalf("failed to write sub.rs: %v", err)
	}

	// 3. Test line counting across Cargo project
	mainPath := filepath.Join(srcDir, "main.rs")
	lines := CountLines(mainPath)
	if lines < 6 {
		t.Errorf("expected at least 6 lines across Cargo files, got %d", lines)
	}

	// 4. Test Cargo Build
	res := Build(mainPath)
	if !res.Success {
		t.Fatalf("Cargo build failed: %s", res.RawOutput)
	}

	// 5. Test Run
	runRes := Run(res.BinaryPath)
	if !runRes.Completed || runRes.ExitCode != 0 {
		t.Errorf("run failed: %v", runRes)
	}
	if !strings.Contains(runRes.Output, "TURBO RUST CARGO OK") {
		t.Errorf("expected 'TURBO RUST CARGO OK', got %q", runRes.Output)
	}
}

func TestRustMultiFileModuleBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create main.rs and foo.rs in same directory (no Cargo.toml)
	mainCode := `mod foo;
fn main() {
    println!("{}", foo::answer());
}
`
	mainPath := filepath.Join(tmpDir, "main.rs")
	if err := os.WriteFile(mainPath, []byte(mainCode), 0644); err != nil {
		t.Fatalf("failed to write main.rs: %v", err)
	}

	fooCode := `pub fn answer() -> i32 {
    42
}
`
	fooPath := filepath.Join(tmpDir, "foo.rs")
	if err := os.WriteFile(fooPath, []byte(fooCode), 0644); err != nil {
		t.Fatalf("failed to write foo.rs: %v", err)
	}

	// 2. Count lines
	lines := CountLines(mainPath)
	if lines < 6 {
		t.Errorf("expected at least 6 lines across both files, got %d", lines)
	}

	// 3. Build with standalone rustc
	res := Build(mainPath)
	if !res.Success {
		t.Fatalf("rustc module build failed: %s", res.RawOutput)
	}

	// 4. Run
	runRes := Run(res.BinaryPath)
	if !runRes.Completed || runRes.ExitCode != 0 {
		t.Errorf("run failed: %v", runRes)
	}
	if !strings.Contains(runRes.Output, "42") {
		t.Errorf("expected '42', got %q", runRes.Output)
	}
}

