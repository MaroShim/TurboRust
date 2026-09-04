package compiler

import (
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

