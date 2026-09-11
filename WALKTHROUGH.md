# Turbo Rust (`tr`) LLDB/GDB 디버거 연동 및 자동 폴백 구현 결과

## 1. 개요
Turbo Rust (`tr`)에 시스템에 설치된 네이티브 디버거(**`rust-lldb`**, **`lldb`**, 또는 **`gdb`**)를 감지하여 실제 컴파일된 바이너리를 하드웨어 수준에서 디버깅하고, 디버거가 없는 환경에서는 기존의 내장 AST 시뮬레이션 디버거(**`RustEngine`**)로 자동 폴백(Fallback)하는 하이브리드 디버거 시스템을 구현하였습니다.

---

## 2. 변경 내용

### 1) [external.go](file:///Users/maro/Projects/tr/internal/debugger/external.go) [신규 추가]
- **`ExternalSession` 구조체**:
  - `exec.Command`로 `rust-lldb`, `lldb`, `gdb` 프로세스 실행 및 `stdin`/`stdout` 파이프 연결.
  - LLDB 및 GDB 명령 프로토콜 지원:
    - LLDB: `settings set auto-confirm true`, `breakpoint set -f <file> -l <line>`, `run`, `thread step-over`, `thread step-in`, `thread continue`, `frame variable`.
    - GDB: `set pagination off`, `set confirm off`, `break <file>:<line>`, `run`, `next`, `step`, `continue`, `info locals`.
  - **정지 위치 정규식 파싱**: `frame #0: 0x... func at file.rs:line:col` ➔ 에디터의 소스 파일, 줄 번호, 함수명과 연동.
  - **로컬 변수 정규식 파싱**: `(type) name = val` (LLDB) 및 `name = val` (GDB) ➔ **Watches Window**에 변수명, 타입, 값 실시간 반영.
  - **프로그램 출력 캡처**: 디버깅 대상 프로그램이 출력하는 콘솔 메시지(`println!`)를 버퍼에 기록하여 `Alt+F5` **User Screen**에서 확인 가능.
  - **프로세스 종료 및 리소스 정리**: `Stop()` 호출 시 `process kill` 및 `quit`으로 자식 프로세스 안전 종료.

### 2) [debugger.go](file:///Users/maro/Projects/tr/internal/debugger/debugger.go) [수정]
- `Debugger` 매니저에 `extSession *ExternalSession` 및 `backendType` 필드 추가.
- `StartWithLines()` 호출 시:
  1. `binPath` 바이너리가 존재하고 `FindRustDebugger()`가 `lldb` 또는 `gdb`를 발견하면 `NewExternalSession`을 시작.
  2. 성공 시 네이티브 백엔드(`rust-lldb`, `lldb`, `gdb`)로 세션 운영.
  3. 시작 실패 또는 디버거 미설치 시 기존의 내부 `RustEngine`으로 자동 폴백하여 세션 안정성 보장.
- `Continue()`, `StepOver()`, `StepInto()`, `Stop()`, `GetProgramOutput()` 호출 시 활성화된 백엔드로 투명하게 위임.
- 현재 활성화된 디버거 백엔드 이름을 반환하는 `BackendType() string` 메서드 추가.

### 3) [app.go](file:///Users/maro/Projects/tr/internal/ui/app.go) [수정]
- `StartDebugging()` 시작 시 상태바 메시지에 활성화된 디버거 백엔드 표시 (예: `Debugging started [rust-lldb]` 또는 `Debugging started [internal]`).

### 4) [debugger_test.go](file:///Users/maro/Projects/tr/internal/debugger/debugger_test.go) [수정]
- `TestFindRustDebugger`: 시스템 내 `rust-lldb`/`lldb`/`gdb` 경로 및 타입 감지 단위 테스트.
- `TestInternalDebuggerFallbackWhenNoBinary`: 바이너리가 없거나 존재하지 않는 경로일 때 `internal` 백엔드로 자동 폴백 검증.
- `TestNativeDebuggerExecution`: 실제 `rustc -g` 컴파일 바이너리에 대해 `rust-lldb`로 브레이크포인트 적중, 스텝 오버, 변수 추출(`count = 0`), 프로세스 종료까지 실전 디버깅 사이클 검증.

---

## 3. 검증 결과

- **`tr/internal/debugger` 단위 테스트**:
  ```bash
  === RUN   TestDebuggerBreakpoints
  --- PASS: TestDebuggerBreakpoints (0.00s)
  === RUN   TestDebuggerSession
  --- PASS: TestDebuggerSession (0.00s)
  === RUN   TestRustEngineFibonacci
  --- PASS: TestRustEngineFibonacci (0.00s)
  === RUN   TestRustEngineHello
  --- PASS: TestRustEngineHello (0.00s)
  === RUN   TestFindRustDebugger
      debugger_test.go:188: Detected debugger: path=/Users/maro/.cargo/bin/rust-lldb, type=lldb
  --- PASS: TestFindRustDebugger (0.00s)
  === RUN   TestInternalDebuggerFallbackWhenNoBinary
  --- PASS: TestInternalDebuggerFallbackWhenNoBinary (0.00s)
  === RUN   TestNativeDebuggerExecution
      debugger_test.go:260: Active backend: rust-lldb (path: /Users/maro/.cargo/bin/rust-lldb)
      debugger_test.go:266: Initial stop state: file=.../test_dbg.rs line=1 func=
      debugger_test.go:273: State after StepOver: line=3 vars=[{Name:count Type:int Value:0}]
      debugger_test.go:278: Final state: exited=true exitCode=0
  --- PASS: TestNativeDebuggerExecution (2.32s)
  PASS
  ok  	tr/internal/debugger	2.800s
  ```
- **바이너리 빌드**: `go build -o bin/tr ./cmd/tr` 빌드 성공.
