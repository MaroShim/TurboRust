# Turbo Rust (Version 0.89) 🚀

> **Retro Borland Turbo Pascal / Turbo C Look & Feel IDE for the Rust Language**

**Turbo Rust (`tr`)** is a retro terminal development environment (TUI IDE) that brings together the classic 90s visual interface of Borland's legendary **Turbo Pascal** and **Turbo C** (Turbo Vision signature blue canvas `#0000A8`, double-line frames `╔═╗`, top pull-down menu bar, bottom hotkey bar, and the `Alt+F5` User Screen) with the modern **Rust toolchain (`rustc`, `cargo`)**.

Implemented in Go for instant native performance and zero-dependency terminal rendering with `tcell/v2`.

---

## 📸 Key Features

* **Classic Borland Turbo Vision UI**:
  * Signature Turbo Blue editor canvas (`#0000A8`) with double-line box-drawing characters (`╔═╗`, `║ ║`, `╚═╝`)
  * Text drop shadows and retro window headers (`[■] 1 NONAME00.RS [▲]`)
  * Top pull-down menu bar (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
  * Bottom hotkey bar (`F1 Help`, `F2 Save`, `F3 Open`, `Alt+F9 Compile`, `F9 Make`, `Ctrl+F9 Run`, `Alt+F5 User`, `F10 Menu`)

* **Rust Syntax Highlighting**:
  * Syntax highlighting for Rust keywords (`fn`, `let`, `mut`, `match`, `impl`, `trait`, `struct`, `enum`, `pub`, `use`, `where`, etc.)
  * Built-in primitive and standard types (`i32`, `usize`, `String`, `Option`, `Result`, `Vec`, etc.)
  * Macros (`println!`, `format!`, `vec!`, `panic!`, etc.)
  * Lifetimes (`'a`, `'static`)
  * Literals (strings, raw strings `r#"..."#`, chars, hex/binary/decimal numbers) and comments (`//`, `/* */`)

* **Compiling Modal Dialog & Diagnostics**:
  * Authentic Borland-style "Compiling..." modal dialog displaying target file, total lines, error/warning count, and elapsed build time
  * Automatic `rustc` / `cargo build` integration with compiler error diagnostics (`file.rs:line:col`)
  * **Instant jump to error line and column in the editor** upon build failure

* **Alt+F5 User Screen**:
  * The hallmark Turbo C feature: switch to a full-screen DOS console view to inspect program execution output, and return to the IDE with any keypress

* **Interactive Debugger & Watches Window**:
  * Toggle breakpoints (`●`) with `F4` ➔ highlighted across the entire line with a **solid red bar**
  * `F5` Debug / Continue, `F8` Step Over, `F7` Trace Into
  * Active execution instruction pointer highlighted with a **solid yellow bar (`►`)**
  * Real-time variable inspection (name, type, value) via the bottom **Watches Window** (Debug menu)

* **Borland Retro Sound Effects (Sound FX)**:
  * Simulated authentic 2.5-inch IBM paper-cone PC speaker physics via IIR low-pass filtering and envelope modeling
  * Crisp dual-tone beep on successful compilation; deep error buzz on build failure
  * Sound toggle via `F10` ➔ `Options` ➔ `Sound: ON / OFF`

---

## ⌨️ Keyboard Shortcuts

| Shortcut | Function | Description |
| --- | --- | --- |
| **Alt + F** | **File Menu** | Open File menu directly |
| **Alt + E** | **Edit Menu** | Open Edit menu directly |
| **Alt + S** | **Search Menu** | Open Search menu directly |
| **Alt + R** | **Run Menu** | Open Run menu directly |
| **Alt + C** | **Compile Menu** | Open Compile menu directly |
| **Alt + D** | **Debug Menu** | Open Debug menu directly |
| **Alt + O** | **Options Menu** | Open Options menu directly |
| **Alt + W** | **Window Menu** | Open Window menu directly |
| **Alt + H** | **Help Menu** | Open Help menu directly |
| **F1** | Help / About | Open Turbo Rust info and help dialog |
| **F2** | Save | Save current buffer / Save As |
| **F3** | Open | Open file browser dialog (`.rs` filter) |
| **F4** | **Breakpoint** | Set/unset breakpoint (`●`) on the current line |
| **F5** | **Debug / Continue** | Start debugging / Continue to next breakpoint |
| **F7** | **Trace Into** | Step into function / instruction |
| **F8** | **Step Over** | Step over function / next line |
| **Ctrl + F2** | **Reset Debugger** | Terminate debug session and reset instruction pointer |
| **Ctrl + F** | **Find** | Open find/search dialog |
| **Ctrl + L** | **Search Again** | Find next occurrence |
| **Alt + G** | **Go to Line** | Jump cursor to specific line number (`Ctrl+G`) |
| **Alt + L** | **Line Numbers** | Toggle line number gutter on/off (`F6`) |
| **Ctrl + F9** | **Run** | Build, execute, and display output in the **User Screen** |
| **Alt + F9** | **Compile** | Build with the "Compiling..." statistics modal |
| **F9** | Make | Execute build (`rustc` / `cargo`) |
| **Alt + F5** | **User Screen** | Toggle program execution output screen |
| **F10** | Menu Bar | Focus top pull-down menu bar |
| **Alt + X** | Exit | Quit Turbo Rust |
| **Shift + Arrow Keys** | **Select Block** | Select/highlight text block |
| **Ctrl + Ins** | **Copy** | Copy selected block to clipboard (`Edit ➔ Copy`) |
| **Shift + Del** | **Cut** | Cut selected block (`Edit ➔ Cut`) |
| **Shift + Ins** | **Paste** | Paste clipboard contents at cursor (`Edit ➔ Paste`) |
| **Esc** | Close | Close active modal/dialog or clear selection |

---

## 🛠️ Build & Run

### 1. Build and run Turbo Rust binary

```bash
go build -o bin/tr.exe ./cmd/tr
./bin/tr.exe
```

### 2. Open a specific Rust file

```bash
./bin/tr.exe examples/hello.rs
# or
./bin/tr.exe examples/fibonacci.rs
```

---

# Turbo Rust (Version 0.89) 🚀

> **볼랜드 Turbo Pascal / Turbo C 레트로 감성의 Rust 전용 TUI IDE**

**Turbo Rust (`tr`)**는 90년대 볼랜드(Borland)의 전설적인 **Turbo Pascal**과 **Turbo C** 특유의 비주얼 인터페이스(시그니처 터보 블루 에디터 창 `#0000A8`, 이중선 프레임 `╔═╗`, 상단 드롭다운 메뉴바, 하단 핫키 바, `Alt+F5` User Screen)에 현대의 **Rust 툴체인(`rustc`, `cargo`)**을 완벽하게 융합한 레트로 터미널 개발 환경(TUI IDE)입니다.

Go 언어와 `tcell/v2`로 구현되어 단일 실행 파일로 즉시 실행되며, 깜빡임 없는 터미널 렌더링을 제공합니다.

---

## 📸 주요 특징

- **Classic Borland Turbo Vision UI**:
  - 시그니처 터보 블루 에디터 캔버스 (`#0000A8`) 및 이중선 박스 드로잉 (`╔═╗`, `║ ║`, `╚═╝`)
  - 입체 텍스트 그림자(Drop Shadow) 및 레트로 윈도우 헤더 (`[■] 1 NONAME00.RS [▲]`)
  - 상단 풀다운 메뉴바 (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
  - 하단 단축키 바 (`F1 Help`, `F2 Save`, `F3 Open`, `Alt+F9 Compile`, `F9 Make`, `Ctrl+F9 Run`, `Alt+F5 User`, `F10 Menu`)
- **Rust Syntax Highlighting**:
  - Rust 2021/2024 키워드, 타입(`i32`, `String`, `Option`, `Result` 등), 매크로(`println!`), 라이프타임(`'a`), 리터럴, 주석 구문 강조
- **Compiling 모달 다이얼로그 & 컴파일 에러 점프**:
  - 컴파일 시 볼랜드 특유의 "Compiling..." 팝업 박스(대상 파일, 총 라인 수, 에러/워닝 수, 빌드 경과 시간)
  - 컴파일 에러 발생 시 파일명/라인/에러 목록 표시 및 **에디터 해당 줄과 열로 즉시 점프**
- **Alt+F5 User Screen**:
  - Turbo C의 상징적인 기능! 빌드 실행 결과를 별도의 전체화면 DOS 콘솔 뷰어로 전환하여 확인하고, 아무 키나 누르면 다시 Turbo Rust IDE로 복귀
- **인터랙티브 디버거 & Watches 창**:
  - `F4`로 브레이크포인트(`●`) 설정/해제 ➔ **한 줄 전체 빨간색 바(`Red`)** 표시
  - `F5` 디버깅 시작 / Continue, `F8` Step Over, `F7` Trace Into
  * 실행 중 현재 멈춘 라인은 **한 줄 전체 노란색 바(`Yellow` `►`)**로 시선 집중
  * 하단 **Watches 윈도우**(Debug 메뉴)를 통해 변수명, 타입, 값을 실시간 감시
- **볼랜드 레트로 사운드 이펙트 (Sound FX)**:
  - 컴파일 성공 시 경쾌한 2단 비프음, 컴파일 에러 시 묵직한 에러음
  - 디버거 브레이크포인트 적중 시 피에조 클릭 사운드

---

## ⌨️ 단축키 안내

| 단축키 | 기능 | 설명 |
|---|---|---|
| **Alt + F** | **File Menu** | File 메뉴 즉시 열기 |
| **Alt + E** | **Edit Menu** | Edit 메뉴 즉시 열기 |
| **Alt + S** | **Search Menu** | Search 메뉴 즉시 열기 |
| **Alt + R** | **Run Menu** | Run 메뉴 즉시 열기 |
| **Alt + C** | **Compile Menu** | Compile 메뉴 즉시 열기 |
| **Alt + D** | **Debug Menu** | Debug 메뉴 즉시 열기 |
| **Alt + O** | **Options Menu** | Options 메뉴 즉시 열기 |
| **Alt + W** | **Window Menu** | Window 메뉴 즉시 열기 |
| **Alt + H** | **Help Menu** | Help 메뉴 즉시 열기 |
| **F1** | Help / About | Turbo Rust 정보 및 도움말 대화상자 |
| **F2** | Save | 현재 버퍼 저장 / 다른 이름으로 저장 |
| **F3** | Open | 파일 브라우저 다이얼로그 열기 (`.rs` 필터링) |
| **F4** | **Breakpoint** | 현재 라인 브레이크포인트(`●`) 설정/해제 |
| **F5** | **Debug / Continue** | 디버깅 시작 / 다음 브레이크포인트까지 계속 실행 |
| **F7** | **Trace Into** | 한 줄씩 실행 (함수/명령 진입) |
| **F8** | **Step Over** | 한 줄씩 실행 (함수 건너뛰기) |
| **Ctrl + F2** | **Reset Debugger** | 디버깅 세션 종료 및 실행 포인터 초기화 |
| **Ctrl + F** | **Find** | 문자열 검색 다이얼로그 열기 |
| **Ctrl + L** | **Search Again** | 이전 검색어로 다음 위치 계속 찾기 (Find Next) |
| **Alt + G** | **Go to Line** | 특정 줄 번호로 커서 이동 (`Ctrl+G`) |
| **Alt + L** | **Line Numbers** | 왼쪽 줄 번호 표시 On / Off 토글 (`F6`) |
| **Ctrl + F9** | **Run** | 빌드 후 일반 실행 및 결과 화면(**User Screen**) 표시 |
| **Alt + F9** | **Compile** | "Compiling..." 통계 팝업과 함께 빌드 실행 |
| **F9** | Make | 빌드 실행 (`rustc` / `cargo`) |
| **Alt + F5** | **User Screen** | 프로그램 실행 결과 화면 토글 |
| **F10** | Menu Bar | 상단 풀다운 메뉴바 포커스 토글 |
| **Alt + X** | Exit | Turbo Rust 종료 |
| **Shift + 방향키** | **Select Block** | 텍스트 영역 블록 선택 (하이라이트) |
| **Ctrl + Ins** | **Copy** | 선택한 블록 클립보드에 복사 (`Edit ➔ Copy`) |
| **Shift + Del** | **Cut** | 선택한 블록 잘라내기 (`Edit ➔ Cut`) |
| **Shift + Ins** | **Paste** | 클립보드 내용 커서 위치에 붙여넣기 (`Edit ➔ Paste`) |
| **Esc** | Close | 활성 메뉴/팝업 다이얼로그 닫기, 선택 해제 |

---

## 🛠️ 실행 및 빌드 방법

### 1. 바이너리 빌드 및 실행
```bash
go build -o bin/tr.exe ./cmd/tr
./bin/tr.exe
```

### 2. 특정 Rust 파일 열기
```bash
./bin/tr.exe examples/hello.rs
# 또는
./bin/tr.exe examples/fibonacci.rs
```
