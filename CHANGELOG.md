# Changelog

All notable changes to **Turbo Rust (`tr`)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.90] - 2026-09-21

### Added
- **Submenu Mnemonic Hotkeys**: Direct single-letter execution in drop-down menus (e.g. `File` ➔ `N` New, `O` Open, `S` Save, `A` Save As; `Edit` ➔ `U` Undo, `R` Redo, `C` Copy).
- **Word-by-Word Navigation**: Move cursor word-by-word with `Ctrl + Left/Right` (and macOS `Option + Left/Right`).
- **macOS Option Meta Key Guide**: Added documentation and tips for configuring Option as Meta key in macOS Terminal.app and iTerm2.
- **MIT License**: Added official open-source MIT License.
- **Language-Separated Documentation**: Split `README.md` into dedicated English (`README.md`) and Korean (`README.ko.md`) with reciprocal navigation links.
- **Authentic Terminal Screenshots**: Embedded high-resolution native terminal screenshots in documentation.
- **Windows Defender Notice**: Added guidance for Windows SmartScreen false-positive warnings on unsigned binaries.

## [0.89] - 2026-09-18

### Added
- **Automated Multi-Platform Release CI/CD**: GitHub Actions workflow and `scripts/build_release.sh` building pure static binaries for macOS (Apple Silicon), Linux (amd64), and Windows (amd64) with SHA256 checksums.
- **Mouse Support**: Mouse click, drag block selection, and mouse wheel scrolling across editor, menu bar, and dialogs.
- **Unsaved Changes Confirmation**: Borland-style modal alert dialog prompting to save or discard changes when opening files or exiting.
- **Scratch Buffer Support**: Untitled scratch buffer enabling autocomplete, hover, and syntax styling before saving to disk.
- **User Screen Live Streaming**: Stream live subprocess output to the `Alt+F5` User Screen buffer during debugger sessions.

### Fixed
- **Debugger Stability**: Ensured GDB runs synchronously with `target-async off`, fixed frame emission on steps, and prevented LLDB on macOS from stepping into internal Rust standard library frames.

## [0.88] - 2026-09-14

### Added
- **LSP Code Completion Popup**: Real-time identifier autocompletion popup with authentic Turbo Vision double-line styling.
- **Syntax Color Enhancement**: Refined Rust keyword, type, and macro styling matching VS Code Go Dark+ palette.

### Fixed
- **User Screen Live Streaming**: Stream live program stdout/stderr to `Alt+F5` User Screen during active native debugging sessions.

## [0.80] - 2026-09-11

### Added
- **Code Navigation & Search**: Project-wide text search and `F12` Go to Definition with dedicated search results modal dialog.
- **Navigation History Stack**: Multi-file jump history with bidirectional jump navigation.
- **Multi-Level Undo/Redo**: Full buffer undo and redo stack supporting multi-step rollbacks.
- **LSP Client Integration**: Native `rust-analyzer` client with server connection status badge and hover documentation snippets.
- **Native Debugger Integration**: Multi-backend native debugging with GDB (Linux) and LLDB (macOS).
- **Modular Example Suite**: Multi-file calculation suite example (`main.rs`, `math.rs`, `stats.rs`).

### Refactored
- **Atomic File Safety**: Temporary swap file staging, atomic rename, automatic UTF-8 BOM stripping, and non-text binary file rejection.
- **Native Terminal Cursor**: Centralized `DrawInputField` helper and native terminal cursor rendering.
- **macOS Alt Key Dispatch**: Streamlined Alt/Meta key handling.

### Fixed
- **GDB Synchronization**: Configured GDB with `target-async off` and ensured synchronous frame emission on steps.
- **LLDB Rust Stdlib Filtering**: Prevented LLDB on macOS from stepping into internal Rust standard library frames.
- **Multi-File Stepping**: Resolved multi-file breakpoint tracking and automatic active file switching.

## [0.50] - 2026-09-08

### Added
- **Multi-File Rust Compilation**: Automatic detection of `Cargo.toml` projects (`cargo build`) and standalone multi-module files (`mod foo;`).
- **Cross-Platform Clipboard**: Unified clipboard abstraction supporting system clipboard (`Ctrl+C`, `Ctrl+X`, `Ctrl+V`, `Ctrl+A`).

## [0.10] - 2026-09-04

### Added
- **Classic Borland Turbo Vision UI**: Signature Turbo Blue canvas (`#0000A8`), double-line box frames (`╔═╗`), 3D text drop shadows, top pull-down menu bar (`F10`), and bottom hotkey bar.
- **Rust Code Editor**: Full syntax highlighting for Rust 2021/2024 keywords (`fn`, `let`, `mut`, `match`, `impl`, `trait`, `struct`, `enum`), types, macros (`println!`), lifetimes (`'a`), and comments.
- **Compiling Modal Dialog**: Authentic statistics modal showing target file, total lines, error counts, and elapsed build time with instant jump to error line and column.
- **Alt+F5 User Screen**: Dedicated full-screen console viewer to inspect program stdout/stderr and exit codes.
- **Interactive Debugger**: Breakpoint toggling (`F4`) with red highlight bar, active instruction pointer (`►`) with yellow highlight bar, Step Over (`F8`), Trace Into (`F7`), and bottom **Watches Window** for real-time variable inspection.
- **Retro PC Speaker Sound Effects**: Dual-tone compilation success chime, failure buzz, and debugger step pings with audio toggle (`Options ➔ Sound`).
- **Automated Test Suite**: Unit tests for editor operations, UI components, syntax highlighting, and compiler runner.
