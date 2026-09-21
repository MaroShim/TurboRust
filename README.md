[한국어](./README.ko.md) | [English](./README.md)

# Turbo Rust (Version 0.90)

> **Retro Borland Turbo Pascal / Turbo C Look & Feel IDE for the Rust Language**

**Turbo Rust (`tr`)** is a retro terminal development environment (TUI IDE) that brings together the classic 90s visual interface of Borland's legendary **Turbo Pascal** and **Turbo C** (Turbo Vision signature blue canvas `#0000A8`, double-line frames `╔═╗`, top pull-down menu bar, bottom hotkey bar, and the `Alt+F5` User Screen) with the modern **Rust toolchain (`rustc`, `cargo`)**.

Implemented in Go for instant native performance and zero-dependency terminal rendering with `tcell/v2`.

---

<p align="center">
  <img src="docs/images/screenshot.png" alt="Turbo Rust Screenshot" width="850">
</p>

---

## Key Features

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

* **Compiling Modal Dialog & Cargo / Multi-File Project Support**:
  * Authentic Borland-style "Compiling..." modal dialog displaying target file, total lines, error/warning count, and elapsed build time
  * **Cargo & Multi-File Module Support**: Automatically detects `Cargo.toml` projects to run `cargo build`, parses output binary from `target/debug/`, and compiles standalone `mod foo;` multi-file projects seamlessly
  * **Instant jump to error line and column in the editor** (even across different files) upon build failure

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

## Keyboard Shortcuts

| Shortcut | Function | Description |
|---|---|---|
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
| **Ctrl + Left / Right** | **Word Jump** | Move cursor word-by-word (macOS: **Option + Left / Right**) |
| **Shift + Arrow Keys** | **Select Block** | Select/highlight text block (supports Ctrl/Option for word selection) |
| **Ctrl + A** | **Select All** | Select all buffer text (`Edit ➔ Select All`) |
| **Ctrl + C** / **Ctrl + Ins** | **Copy** | Copy selected block to clipboard (`Edit ➔ Copy`) |
| **Ctrl + X** / **Shift + Del** | **Cut** | Cut selected block to clipboard (`Edit ➔ Cut`) |
| **Ctrl + V** / **Shift + Ins** | **Paste** | Paste clipboard contents at cursor (`Edit ➔ Paste`) |
| **Esc** | Close | Close active modal/dialog or clear selection |

---

## Installation & Build

### 1. Pre-built Binaries (GitHub Releases)

Download ready-to-use standalone executables for your platform from [GitHub Releases](https://github.com/MaroShim/TurboRust/releases):
* **macOS**: `tr-v0.90-darwin-arm64.tar.gz` (Apple Silicon M-series)
* **Linux**: `tr-v0.90-linux-amd64.tar.gz` (64-bit)
* **Windows**: `tr-v0.90-windows-amd64.zip` (64-bit)

> [!NOTE]
> **Windows Defender / SmartScreen Notice**:
> Since these open-source binaries are newly compiled without expensive commercial code-signing certificates, Windows Defender or SmartScreen may occasionally flag them as unrecognized or a false positive.
> If a Windows SmartScreen popup appears, click **"More info" ➔ "Run anyway"** (추가 정보 ➔ 실행) or add an exclusion to run safely. You can also build directly from source using the Go compiler below.

### 2. Install directly via Go (Recommended)

```bash
go install github.com/MaroShim/TurboRust/cmd/tr@latest
```

Ensure `$GOPATH/bin` (or `~/go/bin`) is in your `$PATH`. You can then launch `tr` from anywhere:

```bash
tr
# or open a file directly:
tr examples/hello.rs
```

### 3. Build from Source

```bash
git clone https://github.com/MaroShim/TurboRust.git
cd TurboRust
go build -o bin/tr ./cmd/tr
./bin/tr
```
