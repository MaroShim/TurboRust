package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"tr/internal/compiler"
	"tr/internal/lsp"
	"tr/internal/sound"
	"tr/internal/ui"
	"tr/internal/ui/dialogs"
)

func main() {
	var initialFile string
	if len(os.Args) > 1 {
		initialFile = os.Args[1]
	}

	app, err := ui.NewApp(initialFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Turbo Rust: %v\n", err)
		os.Exit(1)
	}
	defer app.Stop()

	// Initialize dialogs
	compileDlg := dialogs.NewCompileDialog()
	errListDlg := dialogs.NewErrorListDialog()
	openDlg := dialogs.NewOpenFileDialog()
	saveDlg := dialogs.NewSaveFileDialog()
	aboutDlg := dialogs.NewAboutDialog()
	gotoDlg := dialogs.NewGotoLineDialog()
	findDlg := dialogs.NewFindDialog()
	searchResDlg := dialogs.NewSearchResultsDialog()

	app.SetDialogs(compileDlg, errListDlg, openDlg, saveDlg, aboutDlg, gotoDlg)
	app.SetFindDialog(findDlg)
	app.SetSearchResultsDialog(searchResDlg)

	screen := app.Screen()
	editor := app.GetEditor()
	userScreen := app.GetUserScreen()

	// Action dispatcher
	var dispatchAction func(actionID string)
	dispatchAction = func(actionID string) {
		switch actionID {
		case "file_new":
			*editor = *ui.NewEditor("", editor.WindowNumber)
		case "file_open":
			openDlg.Show(".", func(path string) {
				if err := editor.LoadFile(path); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Error opening " + filepath.Base(path) + ": " + err.Error())
				} else {
					app.SetStatusMessage("Opened " + editor.FileName)
					if lspClient := app.GetLSP(); lspClient != nil && lspClient.IsAvailable() {
						_ = lspClient.DidOpen(editor.FilePath, strings.Join(editor.Lines, "\n"))
					}
				}
			})
		case "file_save":
			if editor.FilePath == "" || editor.FilePath == "NONAME00.RS" {
				saveDlg.Show("main.rs", func(path string) {
					if err := editor.SaveAs(path); err != nil {
						sound.PlayError()
						app.SetStatusMessage("Error saving " + filepath.Base(path) + ": " + err.Error())
					} else {
						sound.PlaySuccess()
						app.SetStatusMessage("Saved " + editor.FileName)
						if lspClient := app.GetLSP(); lspClient != nil && lspClient.IsAvailable() {
							_ = lspClient.DidOpen(editor.FilePath, strings.Join(editor.Lines, "\n"))
						}
					}
				})
			} else {
				if err := editor.SaveFile(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Error saving " + editor.FileName + ": " + err.Error())
				} else {
					sound.PlaySuccess()
					app.SetStatusMessage("Saved " + editor.FileName)
					if lspClient := app.GetLSP(); lspClient != nil && lspClient.IsAvailable() {
						_ = lspClient.DidChange(editor.FilePath, strings.Join(editor.Lines, "\n"))
					}
				}
			}
		case "file_save_as":
			defaultName := editor.FileName
			if defaultName == "" || defaultName == "NONAME00.RS" {
				defaultName = "main.rs"
			}
			saveDlg.Show(defaultName, func(path string) {
				if err := editor.SaveAs(path); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Error saving " + filepath.Base(path) + ": " + err.Error())
				} else {
					sound.PlaySuccess()
					app.SetStatusMessage("Saved " + editor.FileName)
					if lspClient := app.GetLSP(); lspClient != nil && lspClient.IsAvailable() {
						_ = lspClient.DidOpen(editor.FilePath, strings.Join(editor.Lines, "\n"))
					}
				}
			})
		case "app_exit":
			app.Stop()
			os.Exit(0)
		case "run_run":
			bRes, _ := app.RunCurrent()
			if !bRes.Success {
				sound.PlayError()
				compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
			} else {
				sound.PlaySuccess()
			}
		case "run_userscreen":
			userScreen.Show()
		case "compile_compile", "compile_make", "compile_buildall":
			bRes := app.CompileCurrent()
			if bRes.Success {
				sound.PlaySuccess()
			} else {
				sound.PlayError()
			}
			compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
		case "debug_continue":
			if !app.GetDebugger().IsActive() {
				bRes, err := app.StartDebugging()
				if err != nil && bRes != nil && !bRes.Success {
					sound.PlayError()
					compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
				} else {
					sound.PlayBreakpoint()
					// If no breakpoints were set in editor, run through to completion
					hasBPs := false
					for _, set := range editor.Breakpoints {
						if set {
							hasBPs = true
							break
						}
					}
					if !hasBPs && app.GetDebugger().IsActive() {
						if err := app.DebugContinue(); err != nil {
							sound.PlayError()
							app.SetStatusMessage("Debug error: " + err.Error())
						}
					}
				}
			} else {
				if err := app.DebugContinue(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Debug error: " + err.Error())
				} else {
					sound.PlayBreakpoint()
				}
			}
		case "debug_step_over":
			if !app.GetDebugger().IsActive() {
				bRes, err := app.StartDebugging()
				if err != nil && bRes != nil && !bRes.Success {
					sound.PlayError()
					compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
				} else {
					sound.PlayBreakpoint()
				}
			} else {
				if err := app.DebugStepOver(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Debug error: " + err.Error())
				} else {
					sound.PlayBreakpoint()
				}
			}
		case "debug_step_into":
			if !app.GetDebugger().IsActive() {
				bRes, err := app.StartDebugging()
				if err != nil && bRes != nil && !bRes.Success {
					sound.PlayError()
					compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
				} else {
					sound.PlayBreakpoint()
				}
			} else {
				if err := app.DebugStepInto(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Debug error: " + err.Error())
				} else {
					sound.PlayBreakpoint()
				}
			}
		case "debug_stop":
			app.StopDebugging()
		case "debug_watches":
			watch := app.GetWatchWindow()
			watch.Visible = !watch.Visible
		case "debug_toggle_bp":
			currLine := editor.CursorY + 1
			app.ToggleBreakpoint(currLine)
			sound.PlayBell()
		case "options_toggle_linenums":
			editor.ToggleLineNumbers()
		case "options_toggle_sound":
			en := sound.Toggle()
			app.GetMenuBar().SetSoundEnabled(en)
		case "search_find":
			initQ := editor.GetWordUnderCursor()
			if initQ == "" {
				initQ = editor.LastFindQuery
			}
			findDlg.ShowWithTitle("Find", initQ, func(query string, caseSensitive bool) {
				found := editor.FindNext(query, caseSensitive)
				if found {
					sound.PlayBell()
					app.SetStatusMessage(fmt.Sprintf("Found %q", query))
				} else {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("Search string not found: %q", query))
				}
			})
		case "search_project":
			initQ := editor.GetWordUnderCursor()
			if initQ == "" {
				initQ = editor.LastFindQuery
			}
			findDlg.ShowWithTitle("Find in Project", initQ, func(query string, caseSensitive bool) {
				matches := compiler.SearchInProject(editor.FilePath, query, caseSensitive)
				if len(matches) > 0 {
					sound.PlayBell()
					rootDir := compiler.GetSearchRootDir(editor.FilePath)
					searchResDlg.Show(matches, rootDir, func(match compiler.SearchMatch) {
						editor.PushNavLocation()
						if match.File != "" && match.File != editor.FilePath {
							if err := editor.LoadFile(match.File); err != nil {
								sound.PlayError()
								app.SetStatusMessage("Failed to open " + filepath.Base(match.File) + ": " + err.Error())
								return
							}
						}
						editor.GotoLine(match.Line, match.Column)
						app.SetStatusMessage(fmt.Sprintf("Jumped to %s:%d", filepath.Base(match.File), match.Line))
					})
				} else {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("No matches found for %q in project", query))
				}
			})
		case "search_definition":
			// 1. Try LSP definition first if available
			jumped := false
			if lspClient := app.GetLSP(); lspClient != nil && lspClient.IsAvailable() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				targetFile, targetLine, targetCol, ok := lspClient.Definition(ctx, editor.FilePath, editor.CursorY, editor.CursorX)
				cancel()
				if ok && targetFile != "" {
					sound.PlayBell()
					editor.PushNavLocation()
					if targetFile != editor.FilePath {
						if err := editor.LoadFile(targetFile); err != nil {
							sound.PlayError()
							app.SetStatusMessage("Failed to open " + filepath.Base(targetFile) + ": " + err.Error())
							return
						}
					}
					editor.GotoLine(targetLine, targetCol)
					sym := editor.GetWordUnderCursor()
					if sym != "" {
						app.SetStatusMessage(fmt.Sprintf("Jumped to definition of %q (%s:%d)", sym, filepath.Base(targetFile), targetLine))
					} else {
						app.SetStatusMessage(fmt.Sprintf("Jumped to %s:%d", filepath.Base(targetFile), targetLine))
					}
					jumped = true
				}
			}

			// 2. Fall back to built-in AST / regex project scanner
			if !jumped {
				sym := editor.GetWordUnderCursor()
				if sym == "" {
					sym = editor.LastFindQuery
				}
				if sym != "" {
					file, line, col, ok := compiler.FindDefinitionInProject(editor.FilePath, sym)
					if ok {
						sound.PlayBell()
						editor.PushNavLocation()
						if file != "" && file != editor.FilePath {
							if err := editor.LoadFile(file); err != nil {
								sound.PlayError()
								app.SetStatusMessage("Failed to open " + filepath.Base(file) + ": " + err.Error())
								return
							}
						}
						editor.GotoLine(line, col)
						app.SetStatusMessage(fmt.Sprintf("Jumped to definition of %q (%s:%d)", sym, filepath.Base(file), line))
					} else {
						sound.PlayError()
						app.SetStatusMessage(fmt.Sprintf("Definition not found for %q", sym))
					}
				} else {
					app.SetStatusMessage("No symbol under cursor (press F12 on function name)")
				}
			}
		case "search_hover":
			lspClient := app.GetLSP()
			if lspClient == nil || !lspClient.IsAvailable() {
				app.SetStatusMessage("Hover requires active LSP server (rust-analyzer)")
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				snippet, ok := lspClient.Hover(ctx, editor.FilePath, editor.CursorY, editor.CursorX)
				cancel()
				if ok && snippet != "" {
					sound.PlayBell()
					app.SetStatusMessage(fmt.Sprintf("Hover: %s", snippet))
				} else {
					sym := editor.GetWordUnderCursor()
					if sym != "" {
						app.SetStatusMessage(fmt.Sprintf("No type info for %q", sym))
					} else {
						app.SetStatusMessage("No hover information available")
					}
				}
			}
		case "options_lsp_status":
			lspClient := app.GetLSP()
			if lspClient != nil && lspClient.IsAvailable() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("LSP Server: %s [Active] Path: %s", lspClient.ServerName(), lspClient.BinPath()))
			} else {
				sound.PlayError()
				if bin, found := lsp.FindRustAnalyzer(); found {
					app.SetStatusMessage(fmt.Sprintf("LSP Server: rust-analyzer found at %s [Inactive/Starting]", bin))
				} else {
					app.SetStatusMessage("LSP Server: rust-analyzer not found. Using built-in AST scanner.")
				}
			}
		case "search_prev_pos":
			if editor.NavigateBack() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Navigated back to %s:%d", editor.FileName, editor.CursorY+1))
			} else {
				app.SetStatusMessage("Navigation history: at oldest location")
			}
		case "search_next_pos":
			if editor.NavigateForward() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Navigated forward to %s:%d", editor.FileName, editor.CursorY+1))
			} else {
				app.SetStatusMessage("Navigation history: at newest location")
			}
		case "edit_undo":
			if editor.Undo() {
				sound.PlayBell()
				app.SetStatusMessage("Undo performed")
			} else {
				app.SetStatusMessage("Already at oldest change")
			}
		case "edit_redo":
			if editor.Redo() {
				sound.PlayBell()
				app.SetStatusMessage("Redo performed")
			} else {
				app.SetStatusMessage("Already at newest change")
			}
		case "search_again":
			if editor.LastFindQuery != "" {
				found := editor.FindNext(editor.LastFindQuery, editor.LastCaseSensitive)
				if found {
					sound.PlayBell()
				} else {
					sound.PlayError()
				}
			} else {
				dispatchAction("search_find")
			}
		case "search_goto":
			gotoDlg.Show(editor.CursorY+1, len(editor.Lines), func(targetLine int) {
				editor.GotoLine(targetLine, 1)
			})
		case "edit_copy":
			txt := editor.GetSelectedText()
			if editor.CopySelection() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Copied %d characters to clipboard", len([]rune(txt))))
			} else {
				app.SetStatusMessage("No text selected to copy (use Shift+Arrows)")
			}
		case "edit_cut":
			txt := editor.GetSelectedText()
			if editor.CutSelection() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Cut %d characters to clipboard", len([]rune(txt))))
			} else {
				app.SetStatusMessage("No text selected to cut (use Shift+Arrows)")
			}
		case "edit_paste":
			clip := ui.GetClipboard()
			if clip != "" {
				editor.PasteText(clip)
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Pasted %d characters from clipboard", len([]rune(clip))))
			} else {
				app.SetStatusMessage("Clipboard is empty")
			}
		case "edit_clear":
			if editor.DeleteSelection() {
				app.SetStatusMessage("Selection deleted")
			} else {
				app.SetStatusMessage("No text selected to clear")
			}
		case "edit_select_all":
			editor.SelectAll()
			app.SetStatusMessage("All text selected")
		case "help_about":
			aboutDlg.Show()
		}
	}

	app.SetActionHandler(dispatchAction)

	// Main event loop
	for {
		app.Redraw()

		ev := screen.PollEvent()
		switch tev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()

		case *tcell.EventKey:
			key := tev.Key()
			mod := tev.Modifiers()
			ch := tev.Rune()

			// 1. User Screen handles any key to return to IDE
			if userScreen.Active {
				if key == tcell.KeyUp {
					userScreen.ScrollUp()
				} else if key == tcell.KeyDown {
					_, h := screen.Size()
					userScreen.ScrollDown(h)
				} else {
					userScreen.Hide()
				}
				continue
			}

			// 2. Modals handling
			if compileDlg.Visible {
				if key == tcell.KeyEnter {
					compileDlg.Hide()
					if compileDlg.Result != nil && len(compileDlg.Result.Errors) > 0 {
						errListDlg.Show(compileDlg.Result.Errors, func(errItem compiler.CompileError) {
							if errItem.File != "" && errItem.File != editor.FilePath {
								if err := editor.LoadFile(errItem.File); err != nil {
									sound.PlayError()
									app.SetStatusMessage("Failed to open " + filepath.Base(errItem.File) + ": " + err.Error())
									return
								}
							}
							editor.GotoLine(errItem.Line, errItem.Column)
						})
					}
				} else if key == tcell.KeyEscape || key == tcell.KeyRune {
					compileDlg.Hide()
				}
				continue
			}

			if errListDlg.Visible {
				switch key {
				case tcell.KeyUp:
					errListDlg.MoveUp()
				case tcell.KeyDown:
					errListDlg.MoveDown()
				case tcell.KeyEnter:
					errListDlg.SelectCurrent()
				case tcell.KeyEscape:
					errListDlg.Hide()
				}
				continue
			}

			if openDlg.Visible {
				switch key {
				case tcell.KeyUp:
					openDlg.MoveUp()
				case tcell.KeyDown:
					openDlg.MoveDown()
				case tcell.KeyEnter:
					openDlg.HandleEnter()
				case tcell.KeyEscape:
					openDlg.Hide()
				}
				continue
			}

			if saveDlg.Visible {
				switch key {
				case tcell.KeyEnter:
					saveDlg.Confirm()
				case tcell.KeyEscape:
					saveDlg.Hide()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					saveDlg.Backspace()
				case tcell.KeyRune:
					saveDlg.InsertRune(ch)
				}
				continue
			}

			if aboutDlg.Visible {
				if key == tcell.KeyEnter || key == tcell.KeyEscape || key == tcell.KeyRune {
					aboutDlg.Hide()
				}
				continue
			}

			if gotoDlg.Visible {
				switch key {
				case tcell.KeyEnter:
					gotoDlg.Confirm()
				case tcell.KeyEscape:
					gotoDlg.Hide()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					gotoDlg.Backspace()
				case tcell.KeyRune:
					gotoDlg.InsertRune(ch)
				}
				continue
			}

			if findDlg.Visible {
				switch key {
				case tcell.KeyEnter:
					findDlg.Confirm()
				case tcell.KeyEscape:
					findDlg.Hide()
				case tcell.KeyTab:
					findDlg.NextField()
				case tcell.KeyBacktab:
					findDlg.PrevField()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					findDlg.Backspace()
				case tcell.KeyRune:
					findDlg.InsertRune(ch)
				}
				continue
			}

			if searchResDlg.Visible {
				switch key {
				case tcell.KeyUp:
					searchResDlg.MoveUp()
				case tcell.KeyDown:
					searchResDlg.MoveDown()
				case tcell.KeyEnter:
					searchResDlg.SelectCurrent()
				case tcell.KeyEscape:
					searchResDlg.Hide()
				}
				continue
			}

			// 3. Global Shortcuts (Turbo C / Turbo Pascal Standard Alt combinations)
			isAlt := (mod == tcell.ModAlt)

			if isAlt {
				if key == tcell.KeyF9 {
					// Alt+F9: Compile
					dispatchAction("compile_compile")
					continue
				} else if key == tcell.KeyF5 {
					// Alt+F5: User Screen
					dispatchAction("run_userscreen")
					continue
				} else if key == tcell.KeyF1 {
					// Alt+F1: Hover / Type info
					dispatchAction("search_hover")
					continue
				} else if key == tcell.KeyF3 {
					// Alt+F3: Find in Project
					dispatchAction("search_project")
					continue
				} else if key == tcell.KeyLeft {
					// Alt+Left: Previous Location
					dispatchAction("search_prev_pos")
					continue
				} else if key == tcell.KeyRight {
					// Alt+Right: Next Location
					dispatchAction("search_next_pos")
					continue
				} else if key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
					// Alt+Backspace: Undo (Classic Turbo Vision convention)
					dispatchAction("edit_undo")
					continue
				} else if ch == 'f' || ch == 'F' {
					// Alt+F: Open File Menu
					app.OpenMenuAt(0)
					continue
				} else if ch == 'e' || ch == 'E' {
					// Alt+E: Open Edit Menu
					app.OpenMenuAt(1)
					continue
				} else if ch == 's' || ch == 'S' {
					// Alt+S: Open Search Menu
					app.OpenMenuAt(2)
					continue
				} else if ch == 'r' || ch == 'R' {
					// Alt+R: Open Run Menu
					app.OpenMenuAt(3)
					continue
				} else if ch == 'c' || ch == 'C' {
					// Alt+C: Open Compile Menu
					app.OpenMenuAt(4)
					continue
				} else if ch == 'd' || ch == 'D' {
					// Alt+D: Open Debug Menu
					app.OpenMenuAt(5)
					continue
				} else if ch == 'o' || ch == 'O' {
					// Alt+O: Open Options Menu
					app.OpenMenuAt(6)
					continue
				} else if ch == 'w' || ch == 'W' {
					// Alt+W: Open Window Menu
					app.OpenMenuAt(7)
					continue
				} else if ch == 'h' || ch == 'H' {
					// Alt+H: Open Help Menu
					app.OpenMenuAt(8)
					continue
				} else if ch == 'n' || ch == 'N' {
					// Alt+N: Step Over
					dispatchAction("debug_step_over")
					continue
				} else if ch == 'l' || ch == 'L' {
					// Alt+L: Toggle Line Numbers
					dispatchAction("options_toggle_linenums")
					continue
				} else if ch == 'g' || ch == 'G' {
					// Alt+G: Go to Line
					dispatchAction("search_goto")
					continue
				} else if ch == 'q' || ch == 'Q' {
					// Alt+Q: Stop Debugger
					dispatchAction("debug_stop")
					continue
				} else if ch == 'x' || ch == 'X' {
					// Alt+X: Exit
					dispatchAction("app_exit")
					return
				}
			}

			if mod&tcell.ModCtrl != 0 {
				if key == tcell.KeyCtrlZ {
					// Ctrl+Z: Undo
					dispatchAction("edit_undo")
					continue
				} else if key == tcell.KeyCtrlY {
					// Ctrl+Y: Redo
					dispatchAction("edit_redo")
					continue
				} else if key == tcell.KeyCtrlUnderscore || (ch == '-' && mod&tcell.ModCtrl != 0) {
					if mod&tcell.ModShift != 0 {
						// Ctrl+Shift+-: Next Location
						dispatchAction("search_next_pos")
					} else {
						// Ctrl+-: Previous Location
						dispatchAction("search_prev_pos")
					}
					continue
				} else if key == tcell.KeyCtrlC {
					// Ctrl+C: Copy
					dispatchAction("edit_copy")
					continue
				} else if key == tcell.KeyCtrlX {
					// Ctrl+X: Cut
					dispatchAction("edit_cut")
					continue
				} else if key == tcell.KeyCtrlV {
					// Ctrl+V: Paste
					dispatchAction("edit_paste")
					continue
				} else if key == tcell.KeyCtrlA {
					// Ctrl+A: Select All
					dispatchAction("edit_select_all")
					continue
				} else if key == tcell.KeyF1 {
					// Ctrl+F1: Hover / Type info (laptop fallback for Alt+F1)
					dispatchAction("search_hover")
					continue
				} else if key == tcell.KeyF9 {
					// Ctrl+F9: Run
					dispatchAction("run_run")
					continue
				} else if key == tcell.KeyF2 {
					// Ctrl+F2: Stop Debugger
					dispatchAction("debug_stop")
					continue
				} else if key == tcell.KeyCtrlF {
					// Ctrl+F: Find
					dispatchAction("search_find")
					continue
				} else if key == tcell.KeyCtrlL {
					// Ctrl+L: Search again (Find Next)
					dispatchAction("search_again")
					continue
				} else if key == tcell.KeyCtrlG {
					// Ctrl+G: Go to Line
					dispatchAction("search_goto")
					continue
				} else if key == tcell.KeyInsert {
					// Ctrl+Ins: Copy
					dispatchAction("edit_copy")
					continue
				} else if key == tcell.KeyDelete {
					// Ctrl+Del: Clear
					dispatchAction("edit_clear")
					continue
				}
			}

			// Function keys
			switch key {
			case tcell.KeyF1:
				dispatchAction("help_about")
				continue
			case tcell.KeyF2:
				dispatchAction("file_save")
				continue
			case tcell.KeyF3:
				dispatchAction("file_open")
				continue
			case tcell.KeyF4:
				dispatchAction("debug_toggle_bp")
				continue
			case tcell.KeyF5:
				dispatchAction("debug_continue")
				continue
			case tcell.KeyF6:
				dispatchAction("options_toggle_linenums")
				continue
			case tcell.KeyF7:
				dispatchAction("debug_step_into")
				continue
			case tcell.KeyF8:
				dispatchAction("debug_step_over")
				continue
			case tcell.KeyF9:
				dispatchAction("compile_make")
				continue
			case tcell.KeyF12:
				if mod&tcell.ModShift != 0 {
					// Shift+F12: Previous Location
					dispatchAction("search_prev_pos")
				} else {
					dispatchAction("search_definition")
				}
				continue
			}

			// 4. MenuBar navigation
			// Let's check menu state via app
			// If F10 was pressed or menu is active:
			if key == tcell.KeyF10 {
				app.ToggleMenu()
				continue
			}

			if app.IsMenuActive() {
				switch key {
				case tcell.KeyLeft:
					app.MenuMoveLeft()
				case tcell.KeyRight:
					app.MenuMoveRight()
				case tcell.KeyUp:
					app.MenuMoveUp()
				case tcell.KeyDown:
					app.MenuMoveDown()
				case tcell.KeyEnter:
					act := app.MenuSelect()
					if act != "" {
						dispatchAction(act)
					}
				case tcell.KeyEscape:
					app.MenuClose()
				}
				continue
			}

			// 5. Code Editor Editing Controls
			isShift := (mod&tcell.ModShift != 0)
			if isShift {
				switch key {
				case tcell.KeyLeft:
					editor.StartSelection()
					editor.MoveLeft()
					editor.UpdateSelection()
					continue
				case tcell.KeyRight:
					editor.StartSelection()
					editor.MoveRight()
					editor.UpdateSelection()
					continue
				case tcell.KeyUp:
					editor.StartSelection()
					editor.MoveUp()
					editor.UpdateSelection()
					continue
				case tcell.KeyDown:
					editor.StartSelection()
					editor.MoveDown()
					editor.UpdateSelection()
					continue
				case tcell.KeyHome:
					editor.StartSelection()
					editor.MoveHome()
					editor.UpdateSelection()
					continue
				case tcell.KeyEnd:
					editor.StartSelection()
					editor.MoveEnd()
					editor.UpdateSelection()
					continue
				case tcell.KeyDelete:
					// Shift+Del: Cut
					dispatchAction("edit_cut")
					continue
				case tcell.KeyInsert:
					// Shift+Ins: Paste
					dispatchAction("edit_paste")
					continue
				}
			}


			switch key {
			case tcell.KeyLeft:
				editor.ClearSelection()
				editor.MoveLeft()
			case tcell.KeyRight:
				editor.ClearSelection()
				editor.MoveRight()
			case tcell.KeyUp:
				editor.ClearSelection()
				editor.MoveUp()
			case tcell.KeyDown:
				editor.ClearSelection()
				editor.MoveDown()
			case tcell.KeyHome:
				editor.ClearSelection()
				editor.MoveHome()
			case tcell.KeyEnd:
				editor.ClearSelection()
				editor.MoveEnd()
			case tcell.KeyPgUp:
				editor.ClearSelection()
				_, h := screen.Size()
				editor.PageUp(h - 4)
			case tcell.KeyPgDn:
				editor.ClearSelection()
				_, h := screen.Size()
				editor.PageDown(h - 4)
			case tcell.KeyEnter:
				editor.InsertNewLine()
			case tcell.KeyTab:
				editor.InsertTab()
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				editor.Backspace()
			case tcell.KeyDelete:
				editor.Delete()
			case tcell.KeyRune:
				editor.InsertRune(ch)
			}
		}
	}
}
