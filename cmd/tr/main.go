package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"tr/internal/compiler"
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

	app.SetDialogs(compileDlg, errListDlg, openDlg, saveDlg, aboutDlg, gotoDlg)
	app.SetFindDialog(findDlg)

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
				_ = editor.LoadFile(path)
			})
		case "file_save":
			if editor.FilePath == "" || editor.FilePath == "NONAME00.RS" {
				saveDlg.Show("main.rs", func(path string) {
					_ = editor.SaveAs(path)
				})
			} else {
				_ = editor.SaveFile()
			}
		case "file_save_as":
			defaultName := editor.FileName
			if defaultName == "" {
				defaultName = "main.rs"
			}
			saveDlg.Show(defaultName, func(path string) {
				_ = editor.SaveAs(path)
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
						_ = app.DebugContinue()
					}
				}
			} else {
				_ = app.DebugContinue()
				sound.PlayBreakpoint()
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
				_ = app.DebugStepOver()
				sound.PlayBreakpoint()
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
				_ = app.DebugStepInto()
				sound.PlayBreakpoint()
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
			findDlg.Show(editor.LastFindQuery, func(query string, caseSensitive bool) {
				found := editor.FindNext(query, caseSensitive)
				if found {
					sound.PlayBell()
				} else {
					sound.PlayError()
				}
			})
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

			// 3. Global Shortcuts (Turbo C / Turbo Pascal Standard + macOS Option Key Workarounds)
			isAlt := (mod == tcell.ModAlt)

			// macOS Option key translates characters to special unicode symbols by default:
			// Option+L = '¬', Option+X = '≈', Option+W = '∑', Option+N = '˜', Option+S = 'ß', Option+C = 'ç', Option+Q = 'œ', Option+G = '©', Option+F = 'ƒ'
			if isAlt || ch == '¬' || ch == '≈' || ch == '∑' || ch == '˜' || ch == 'ß' || ch == 'ç' || ch == 'œ' || ch == '©' || ch == 'ƒ' {
				if isAlt && key == tcell.KeyF9 {
					// Alt+F9: Compile
					dispatchAction("compile_compile")
					continue
				} else if isAlt && key == tcell.KeyF5 {
					// Alt+F5: User Screen
					dispatchAction("run_userscreen")
					continue
				} else if ch == 'f' || ch == 'F' || ch == 'ƒ' {
					// Alt+F: Open File Menu
					app.OpenMenuAt(0)
					continue
				} else if ch == 'e' || ch == 'E' {
					// Alt+E: Open Edit Menu
					app.OpenMenuAt(1)
					continue
				} else if ch == 's' || ch == 'S' || ch == 'ß' {
					// Alt+S: Open Search Menu
					app.OpenMenuAt(2)
					continue
				} else if ch == 'r' || ch == 'R' {
					// Alt+R: Open Run Menu
					app.OpenMenuAt(3)
					continue
				} else if ch == 'c' || ch == 'C' || ch == 'ç' {
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
				} else if ch == 'w' || ch == 'W' || ch == '∑' {
					// Alt+W: Open Window Menu
					app.OpenMenuAt(7)
					continue
				} else if ch == 'h' || ch == 'H' {
					// Alt+H: Open Help Menu
					app.OpenMenuAt(8)
					continue
				} else if ch == 'n' || ch == 'N' || ch == '˜' {
					// Alt+N: Step Over
					dispatchAction("debug_step_over")
					continue
				} else if ch == 'l' || ch == 'L' || ch == '¬' {
					// Alt+L: Toggle Line Numbers
					dispatchAction("options_toggle_linenums")
					continue
				} else if ch == 'g' || ch == 'G' || ch == '©' {
					// Alt+G: Go to Line
					dispatchAction("search_goto")
					continue
				} else if ch == 'q' || ch == 'Q' || ch == 'œ' {
					// Alt+Q: Stop Debugger
					dispatchAction("debug_stop")
					continue
				} else if ch == 'x' || ch == 'X' || ch == '≈' {
					// Alt+X: Exit
					dispatchAction("app_exit")
					return
				}
			}

			if mod == tcell.ModCtrl {
				if key == tcell.KeyCtrlC {
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
