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
- **Classic Borland Turbo Vision UI**: Signature Turbo Blue canvas (`#0000A8`), double-line box borders (`╔═╗`), 3D text drop shadows, top pull-down menu bar, and bottom hotkey bar.
- **Rust Syntax Highlighting**: Full highlighting for Rust keywords (`fn`, `let`, `mut`, `match`, `impl`, `trait`, `struct`, `enum`), types, macros (`println!`), lifetimes (`'a`), and comments.
- **Compiling Modal Dialog & Cargo Integration**: Authentic statistics modal showing target, total lines, error counts, and elapsed build time. Auto-detects `Cargo.toml` projects or compiles standalone `mod foo;` files with instant jump to error line and column.
- **Native Debugger Integration**: Multi-backend native debugging with GDB (Linux) and LLDB (macOS). Breakpoints (`●`) with red bars, instruction pointer (`►`) with yellow bars, Step Over (`F8`), Trace Into (`F7`), and bottom **Watches Window** for variable inspection.
- **Retro PC Speaker Sound Effects**: Dual-tone compilation success chime, failure buzz, and debugger step pings with audio toggle (`Options ➔ Sound`).
- **Code Intelligence (LSP)**: `rust-analyzer` integration with real-time autocompletion popup, hover documentation snippets, and semantic tokens.
- **Navigation Stack & Undo**: `F12` Go to Definition with multi-file jump history and multi-step Undo/Redo stack.
- **Atomic File Operations**: Safe atomic saves via swap files, UTF-8 BOM auto-stripping, and non-text binary file protection.
- **Cross-Platform Clipboard**: Seamless integration with system clipboard (`Ctrl+C`, `Ctrl+X`, `Ctrl+V`, `Ctrl+A`).
