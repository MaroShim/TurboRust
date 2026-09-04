package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"tr/internal/compiler"
	"tr/internal/debugger"
)

// DialogHolder interfaces
type Dialog interface {
	Draw(screen tcell.Screen, w, h int)
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

	// Modal Dialogs
	compileDialog   Dialog
	errorListDialog Dialog
	openFileDialog  Dialog
	saveFileDialog  Dialog
	aboutDialog     Dialog
	gotoLineDialog  Dialog
	findDialog      Dialog

	// Callbacks for modal interaction
	onAction func(actionID string)
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

	w, h := s.Size()

	app := &App{
		screen:      s,
		running:     true,
		width:       w,
		height:      h,
		menuBar:     NewMenuBar(),
		statusBar:   NewStatusBar(),
		userScreen:  NewUserScreen(),
		editor:      NewEditor(initialFile, 1),
		debugger:    debugger.NewDebugger(),
		watchWindow: NewWatchWindow(2),
	}

	return app, nil
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

func (a *App) Screen() tcell.Screen {
	return a.screen
}

func (a *App) Stop() {
	a.running = false
	if a.debugger != nil {
		_ = a.debugger.Stop()
	}
	a.screen.Fini()
}

func (a *App) GetWatchWindow() *WatchWindow {
	return a.watchWindow
}

func (a *App) SyncDebuggerState() {
	st := a.debugger.GetState()
	a.watchWindow.SetState(st)
	if st.CurrentLine > 0 {
		a.editor.SetCurrentIP(st.CurrentLine)
	} else {
		a.editor.SetCurrentIP(0)
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

	// Register editor breakpoints into debugger
	hasAnyBP := false
	for l := range a.editor.Breakpoints {
		a.debugger.ToggleBreakpoint(absTarget, l)
		hasAnyBP = true
	}

	curLine := a.editor.CursorY + 1
	if !hasAnyBP {
		if curLine < 1 {
			curLine = 1
		}
		a.debugger.ToggleBreakpoint(absTarget, curLine)
		a.editor.ToggleBreakpoint(curLine)
	}

	err := a.debugger.Start(bRes.BinaryPath, absTarget, curLine)
	if err != nil {
		return bRes, err
	}

	a.watchWindow.Visible = true
	a.SyncDebuggerState()

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

	a.editor.Draw(
		a.screen,
		editorInteriorX,
		editorInteriorY,
		editorInteriorW,
		editorInteriorH,
		!a.menuBar.Active,
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

	// 5. Draw Top MenuBar (row 0)
	a.menuBar.Draw(a.screen, a.width)

	// 6. Draw Bottom StatusBar (row height-1)
	a.statusBar.Draw(a.screen, a.height-1, a.width)

	a.screen.Show()
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
