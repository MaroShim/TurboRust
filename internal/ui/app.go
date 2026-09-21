package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/MaroShim/TurboRust/internal/compiler"
	"github.com/MaroShim/TurboRust/internal/debugger"
	"github.com/MaroShim/TurboRust/internal/lsp"
)

// DialogHolder interfaces
type Dialog interface {
	Draw(screen tcell.Screen, w, h int)
	IsVisible() bool
}

// App is the main Turbo Rust IDE application controller
type App struct {
	screen      tcell.Screen
	running     bool
	width       int
	height      int
	menuBar     *MenuBar
	statusBar   *StatusBar
	userScreen  *UserScreen
	editor      *Editor
	debugger    *debugger.Debugger
	watchWindow *WatchWindow
	lspClient   *lsp.Client
	completionPopup *CompletionPopup

	// Modal Dialogs
	compileDialog   Dialog
	errorListDialog Dialog
	openFileDialog  Dialog
	saveFileDialog  Dialog
	aboutDialog     Dialog
	gotoLineDialog  Dialog
	findDialog      Dialog
	searchResultsDialog Dialog
	confirmSaveDialog   Dialog

	// Callbacks for modal interaction
	onAction func(actionID string)

	workDir    string
	scratchDir string
}

func NewAppWithScreen(s tcell.Screen, initialFile string) *App {
	w, h := s.Size()
	return &App{
		screen:          s,
		running:         true,
		width:           w,
		height:          h,
		menuBar:         NewMenuBar(),
		statusBar:       NewStatusBar(),
		userScreen:      NewUserScreen(),
		editor:          NewEditor(initialFile, 1),
		debugger:        debugger.NewDebugger(),
		watchWindow:     NewWatchWindow(2),
		completionPopup: NewCompletionPopup(),
	}
}

func NewApp(initialFile string) (*App, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create tcell screen: %w", err)
	}

	if err := s.Init(); err != nil {
		return nil, fmt.Errorf("failed to init screen: %w", err)
	}

	s.EnableMouse()
	s.Clear()

	app := NewAppWithScreen(s, initialFile)

	// Initialize LSP Client in background so it doesn't block UI startup
	workDir := "."
	if initialFile != "" {
		workDir = filepath.Dir(initialFile)
	}
	if cargoRoot, hasCargo := compiler.FindCargoRoot(workDir); hasCargo {
		workDir = cargoRoot
	}
	app.workDir = workDir

	if initialFile == "" {
		scratchFile := app.EnsureScratchBuffer()
		app.editor.FilePath = scratchFile
		app.editor.FileName = "NONAME00.RS"
		app.editor.IsUntitled = true
	}

	go func() {
		client, err := lsp.StartRustAnalyzerClient(workDir)
		if err == nil && client != nil {
			app.lspClient = client
			app.statusBar.SetLSPStatus("LSP: rust-analyzer", true)
			app.SetStatusMessage("Turbo Rust ready. LSP: rust-analyzer active [F12: Def, Alt+F1: Hover]")
			if app.editor != nil && app.editor.FilePath != "" {
				_ = client.DidOpen(app.editor.FilePath, strings.Join(app.editor.Lines, "\n"))
				app.RequestSemanticTokens()
			}
		} else {
			app.statusBar.SetLSPStatus("LSP: None", false)
		}
	}()

	return app, nil
}

// EnsureScratchBuffer creates a hermetic scratch directory and shadow main.rs for untitled buffers
func (a *App) EnsureScratchBuffer() string {
	workDir := a.workDir
	if workDir == "" {
		workDir = "."
	}
	scratchDir := filepath.Join(workDir, ".tr_scratch")
	_ = os.MkdirAll(scratchDir, 0755)
	scratchFile := filepath.Join(scratchDir, "main.rs")
	if len(a.editor.Lines) > 0 {
		_ = os.WriteFile(scratchFile, []byte(strings.Join(a.editor.Lines, "\n")), 0644)
	}
	a.scratchDir = scratchDir
	return scratchFile
}

func (a *App) SetDialogs(
	compileDlg Dialog,
	errListDlg Dialog,
	openDlg Dialog,
	saveDlg Dialog,
	aboutDlg Dialog,
	gotoDlg Dialog,
) {
	a.compileDialog = compileDlg
	a.errorListDialog = errListDlg
	a.openFileDialog = openDlg
	a.saveFileDialog = saveDlg
	a.aboutDialog = aboutDlg
	a.gotoLineDialog = gotoDlg
}

func (a *App) SetFindDialog(findDlg Dialog) {
	a.findDialog = findDlg
}

func (a *App) SetSearchResultsDialog(searchResDlg Dialog) {
	a.searchResultsDialog = searchResDlg
}

func (a *App) SetConfirmSaveDialog(dlg Dialog) {
	a.confirmSaveDialog = dlg
}

func (a *App) SetActionHandler(handler func(actionID string)) {
	a.onAction = handler
}

func (a *App) GetEditor() *Editor {
	return a.editor
}

func (a *App) GetUserScreen() *UserScreen {
	return a.userScreen
}

func (a *App) GetDebugger() *debugger.Debugger {
	return a.debugger
}

func (a *App) GetMenuBar() *MenuBar {
	return a.menuBar
}

func (a *App) GetStatusBar() *StatusBar {
	return a.statusBar
}

// GetEditorInteriorBounds returns (interiorX, interiorY, interiorW, interiorH) of editor drawing rectangle
func (a *App) GetEditorInteriorBounds() (int, int, int, int) {
	totalWorkH := a.height - 2
	winX := 0
	winY := 1
	winW := a.width - 2
	if winW < 20 {
		winW = 20
	}

	editorH := totalWorkH
	if a.watchWindow.Visible && totalWorkH >= 12 {
		watchH := totalWorkH / 3
		if watchH < 6 {
			watchH = 6
		}
		editorH = totalWorkH - watchH
	}

	return winX + 1, winY + 1, winW - 2, editorH - 2
}

func (a *App) SetStatusMessage(msg string) {

	a.statusBar.SetMessage(msg)
}

func (a *App) ToggleMenu() {
	if a.menuBar.Active {
		a.menuBar.Close()
	} else {
		a.menuBar.Open()
	}
}

func (a *App) OpenMenuAt(index int) {
	a.menuBar.OpenMenu(index)
}

func (a *App) IsMenuActive() bool {
	return a.menuBar.Active
}

func (a *App) MenuClose() {
	a.menuBar.Close()
}

func (a *App) MenuMoveLeft() {
	a.menuBar.MoveLeft()
}

func (a *App) MenuMoveRight() {
	a.menuBar.MoveRight()
}

func (a *App) MenuMoveUp() {
	a.menuBar.MoveUp()
}

func (a *App) MenuMoveDown() {
	a.menuBar.MoveDown()
}

func (a *App) MenuSelect() string {
	act := a.menuBar.GetSelectedAction()
	a.menuBar.Close()
	return act
}

func (a *App) MenuHandleKey(ch rune) (string, bool) {
	return a.menuBar.HandleKey(ch)
}

func (a *App) Screen() tcell.Screen {
	return a.screen
}

func (a *App) Stop() {
	a.running = false
	if a.debugger != nil {
		a.debugger.Stop()
	}
	if a.lspClient != nil {
		_ = a.lspClient.Close()
	}
	if a.scratchDir != "" {
		_ = os.RemoveAll(a.scratchDir)
	}
	a.screen.Fini()
}

func (a *App) GetLSP() *lsp.Client {
	return a.lspClient
}

func (a *App) GetCompletionPopup() *CompletionPopup {
	return a.completionPopup
}

// RequestSemanticTokens requests semantic tokens for current editor file asynchronously
func (a *App) RequestSemanticTokens() {
	if a.lspClient == nil || !a.lspClient.IsAvailable() || a.editor == nil || a.editor.FilePath == "" {
		return
	}

	filePath := a.editor.FilePath
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		spans, err := a.lspClient.SemanticTokensFull(ctx, filePath)
		if err == nil && len(spans) > 0 {
			if a.editor != nil && a.editor.FilePath == filePath {
				a.editor.SetSemanticTokens(spans)
				if a.screen != nil {
					_ = a.screen.PostEvent(tcell.NewEventInterrupt(nil))
				}
			}
		}
	}()
}

func (a *App) GetWatchWindow() *WatchWindow {
	return a.watchWindow
}

func (a *App) ToggleBreakpoint(line int) bool {
	isSet := a.editor.ToggleBreakpoint(line)
	targetFile := a.editor.FilePath
	if targetFile != "" {
		absTarget, _ := filepath.Abs(targetFile)
		if isSet {
			a.debugger.SetBreakpoint(absTarget, line)
		} else {
			a.debugger.RemoveBreakpoint(absTarget, line)
		}
	}
	return isSet
}

func (a *App) SyncDebuggerState() {
	st := a.debugger.GetState()
	a.watchWindow.SetState(st)
	if st.CurrentLine > 0 && st.Active && !st.Exited {
		targetFile := st.CurrentFile
		fileLoaded := false
		if targetFile != "" {
			resolved := targetFile
			if !filepath.IsAbs(resolved) {
				// 1. Try base name in current editor file's directory
				if a.editor.FilePath != "" {
					baseCand := filepath.Join(filepath.Dir(a.editor.FilePath), filepath.Base(targetFile))
					if _, err := os.Stat(baseCand); err == nil {
						resolved = baseCand
					}
				}
				// 2. Try relative to current editor file's directory
				if !filepath.IsAbs(resolved) && a.editor.FilePath != "" {
					cand := filepath.Join(filepath.Dir(a.editor.FilePath), targetFile)
					if _, err := os.Stat(cand); err == nil {
						resolved = cand
					}
				}
				// 3. Try relative to current working directory
				if !filepath.IsAbs(resolved) {
					if cwd, err := os.Getwd(); err == nil {
						cand := filepath.Join(cwd, targetFile)
						if _, err := os.Stat(cand); err == nil {
							resolved = cand
						}
					}
				}
			}
			if absResolved, err := filepath.Abs(resolved); err == nil {
				resolved = absResolved
			}
			if fi, err := os.Stat(resolved); err == nil && fi.Mode().IsRegular() {
				cleanTarget := filepath.Clean(resolved)
				cleanCurrent := filepath.Clean(a.editor.FilePath)
				if absCur, err := filepath.Abs(cleanCurrent); err == nil {
					cleanCurrent = absCur
				}
				if cleanTarget != cleanCurrent {
					_ = a.editor.LoadFile(cleanTarget)
				}
				fileLoaded = true
			} else if filepath.Base(a.editor.FilePath) == filepath.Base(targetFile) {
				fileLoaded = true
			}
		}
		if fileLoaded {
			a.editor.SetCurrentIP(st.CurrentLine)
		}
	} else {
		a.editor.SetCurrentIP(0)
	}

	if a.userScreen != nil {
		if a.debugger.IsActive() {
			out := a.debugger.GetProgramOutput()
			a.userScreen.SetLiveOutput(out, st.CurrentFile, st.CurrentLine)
		} else if st.Exited {
			out := a.debugger.GetProgramOutput()
			a.userScreen.SetExecutionResult(out, st.ExitCode, "debug")
		}
	}

	// Update status bar items based on debug active state
	if a.debugger.IsActive() {
		a.statusBar.Items = []StatusItem{
			{KeyName: "F5", Desc: "Cont", Action: "debug_continue"},
			{KeyName: "F7", Desc: "Trace", Action: "debug_step_into"},
			{KeyName: "F8", Desc: "Step", Action: "debug_step_over"},
			{KeyName: "Ctrl+F2", Desc: "Reset", Action: "debug_stop"},
			{KeyName: "Alt+F5", Desc: "User", Action: "run_userscreen"},
			{KeyName: "F10", Desc: "Menu", Action: "menu_toggle"},
		}
	} else {
		a.statusBar.Items = []StatusItem{
			{KeyName: "F1", Desc: "Help", Action: "help_about"},
			{KeyName: "F2", Desc: "Save", Action: "file_save"},
			{KeyName: "F3", Desc: "Open", Action: "file_open"},
			{KeyName: "Alt+F9", Desc: "Compile", Action: "compile_compile"},
			{KeyName: "F9", Desc: "Make", Action: "compile_make"},
			{KeyName: "Ctrl+F9", Desc: "Run", Action: "run_run"},
			{KeyName: "Alt+F5", Desc: "User", Action: "run_userscreen"},
			{KeyName: "F10", Desc: "Menu", Action: "menu_toggle"},
		}
	}
}

// StartDebugging compiles with debug symbols and initiates debug session
func (a *App) StartDebugging() (*compiler.BuildResult, error) {
	// Auto-save
	targetFile := a.editor.FilePath
	if targetFile == "" || a.editor.Dirty {
		if targetFile == "" {
			targetFile = filepath.Join(os.TempDir(), "turborust_dbg_main.rs")
			_ = a.editor.SaveAs(targetFile)
		} else {
			_ = a.editor.SaveFile()
		}
	}

	bRes := compiler.Build(targetFile)
	if !bRes.Success {
		return bRes, fmt.Errorf("debug build failed")
	}

	absTarget, _ := filepath.Abs(targetFile)

	// 1. Sync current active editor file breakpoints into FileBreakpoints
	if a.editor.FilePath != "" {
		curKey := filepath.Clean(a.editor.FilePath)
		if len(a.editor.Breakpoints) > 0 {
			saved := make(map[int]bool)
			for k, v := range a.editor.Breakpoints {
				if v {
					saved[k] = true
				}
			}
			if a.editor.FileBreakpoints == nil {
				a.editor.FileBreakpoints = make(map[string]map[int]bool)
			}
			a.editor.FileBreakpoints[curKey] = saved
		} else if a.editor.FileBreakpoints != nil {
			delete(a.editor.FileBreakpoints, curKey)
		}
	}

	// 2. Clear and synchronize all editor breakpoints across ALL files into debugger
	a.debugger.ClearBreakpoints()
	hasAnyBP := false
	if a.editor.FileBreakpoints != nil {
		for f, lines := range a.editor.FileBreakpoints {
			absF, err := filepath.Abs(f)
			if err != nil {
				absF = f
			}
			for l, set := range lines {
				if set {
					a.debugger.SetBreakpoint(absF, l)
					hasAnyBP = true
				}
			}
		}
	}
	for l, set := range a.editor.Breakpoints {
		if set {
			a.debugger.SetBreakpoint(absTarget, l)
			hasAnyBP = true
		}
	}

	curLine := a.editor.CursorY + 1
	if !hasAnyBP && curLine < 1 {
		curLine = 1
	}

	err := a.debugger.StartWithLines(bRes.BinaryPath, absTarget, a.editor.Lines, curLine)
	if err != nil {
		return bRes, err
	}

	a.watchWindow.Visible = true
	a.SyncDebuggerState()
	a.SetStatusMessage(fmt.Sprintf("Debugging started [%s]", a.debugger.BackendType()))

	return bRes, nil
}

func (a *App) DebugContinue() error {
	err := a.debugger.Continue()
	a.SyncDebuggerState()
	return err
}

func (a *App) DebugStepOver() error {
	err := a.debugger.StepOver()
	a.SyncDebuggerState()
	return err
}

func (a *App) DebugStepInto() error {
	err := a.debugger.StepInto()
	a.SyncDebuggerState()
	return err
}

func (a *App) StopDebugging() {
	_ = a.debugger.Stop()
	a.editor.SetCurrentIP(0)
	a.watchWindow.Visible = false
	a.SyncDebuggerState()
}

// Redraw renders the full desktop, active window, menu, status, and modals
func (a *App) Redraw() {
	a.width, a.height = a.screen.Size()

	// If UserScreen is active, render only UserScreen
	if a.userScreen.Active {
		a.userScreen.Draw(a.screen, a.width, a.height)
		a.screen.Show()
		return
	}

	// 1. Draw Desktop Background (Classic Turbo Vision pattern: ░)
	bgStyle := tcell.StyleDefault.Background(ColorDesktopBg).Foreground(ColorDesktopFg)
	for y := 1; y < a.height-1; y++ {
		for x := 0; x < a.width; x++ {
			a.screen.SetContent(x, y, RuneScrollTrack, nil, bgStyle)
		}
	}

	totalWorkH := a.height - 2
	winX := 0
	winY := 1
	winW := a.width - 2
	if winW < 20 {
		winW = 20
	}

	editorH := totalWorkH
	watchH := 0

	if a.watchWindow.Visible && totalWorkH >= 12 {
		watchH = totalWorkH / 3
		if watchH < 6 {
			watchH = 6
		}
		editorH = totalWorkH - watchH
	}

	// Calculate scroll ratio
	scrollRatio := 0.0
	if len(a.editor.Lines) > 1 {
		scrollRatio = float64(a.editor.CursorY) / float64(len(a.editor.Lines)-1)
	}

	title := a.editor.FileName
	if a.editor.Dirty {
		title = "*" + title
	}

	// 2. Window Frame for Editor
	DrawWindowFrame(
		a.screen,
		winX, winY, winW, editorH,
		title,
		a.editor.WindowNumber,
		!a.menuBar.Active,
		a.editor.CursorY+1,
		a.editor.CursorX+1,
		scrollRatio,
	)

	// Editor Content inside frame
	editorInteriorX := winX + 1
	editorInteriorY := winY + 1
	editorInteriorW := winW - 2
	editorInteriorH := editorH - 2

	editorFocused := !a.menuBar.Active && !a.HasModalVisible()

	a.editor.Draw(
		a.screen,
		editorInteriorX,
		editorInteriorY,
		editorInteriorW,
		editorInteriorH,
		editorFocused,
	)

	// 3. Draw Watch Window if visible
	if a.watchWindow.Visible && watchH > 0 {
		watchY := winY + editorH
		a.watchWindow.Draw(a.screen, winX, watchY, winW, watchH, true)
	}

	// 4. Draw Modals if active
	if a.compileDialog != nil {
		a.compileDialog.Draw(a.screen, a.width, a.height)
	}
	if a.errorListDialog != nil {
		a.errorListDialog.Draw(a.screen, a.width, a.height)
	}
	if a.openFileDialog != nil {
		a.openFileDialog.Draw(a.screen, a.width, a.height)
	}
	if a.saveFileDialog != nil {
		a.saveFileDialog.Draw(a.screen, a.width, a.height)
	}
	if a.aboutDialog != nil {
		a.aboutDialog.Draw(a.screen, a.width, a.height)
	}
	if a.gotoLineDialog != nil {
		a.gotoLineDialog.Draw(a.screen, a.width, a.height)
	}
	if a.findDialog != nil {
		a.findDialog.Draw(a.screen, a.width, a.height)
	}
	if a.searchResultsDialog != nil {
		a.searchResultsDialog.Draw(a.screen, a.width, a.height)
	}
	if a.confirmSaveDialog != nil {
		a.confirmSaveDialog.Draw(a.screen, a.width, a.height)
	}

	// 5. Draw Completion Popup if visible (highest floating window priority below menu)
	if a.completionPopup != nil && a.completionPopup.IsVisible() {
		a.completionPopup.Draw(a.screen)
	}

	// 6. Draw Top MenuBar (row 0)
	a.menuBar.Draw(a.screen, a.width)
	if a.menuBar.Active {
		a.screen.HideCursor()
	}

	// 7. Draw Bottom StatusBar (row height-1)
	a.statusBar.Draw(a.screen, a.height-1, a.width)

	a.screen.Show()
}

// HasModalVisible returns true if any modal dialog is currently open
func (a *App) HasModalVisible() bool {
	for _, d := range []Dialog{
		a.compileDialog, a.errorListDialog, a.openFileDialog,
		a.saveFileDialog, a.aboutDialog, a.gotoLineDialog,
		a.findDialog, a.searchResultsDialog, a.confirmSaveDialog,
	} {
		if d != nil && d.IsVisible() {
			return true
		}
	}
	return false
}

// CompileCurrent compiles current buffer
func (a *App) CompileCurrent() *compiler.BuildResult {
	// Auto-save to temp file if not saved or dirty
	targetFile := a.editor.FilePath
	if targetFile == "" || a.editor.Dirty {
		if targetFile == "" {
			targetFile = filepath.Join(os.TempDir(), "turborust_temp_main.rs")
			_ = a.editor.SaveAs(targetFile)
		} else {
			_ = a.editor.SaveFile()
		}
	}

	res := compiler.Build(targetFile)
	return res
}

// RunCurrent builds and runs the current file, capturing output for User Screen
func (a *App) RunCurrent() (*compiler.BuildResult, *compiler.RunResult) {
	bRes := a.CompileCurrent()
	if !bRes.Success {
		return bRes, nil
	}

	rRes := compiler.Run(bRes.BinaryPath)

	durStr := fmt.Sprintf("%.2fs", rRes.Duration.Seconds())
	a.userScreen.SetExecutionResult(rRes.Output, rRes.ExitCode, durStr)
	a.userScreen.Show()

	// Clean up temp binary
	_ = os.Remove(bRes.BinaryPath)

	return bRes, rRes
}
