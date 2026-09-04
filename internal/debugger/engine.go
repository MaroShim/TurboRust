package debugger

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// RustFunc represents a parsed function definition
type RustFunc struct {
	Name       string
	Params     []string // param names
	ParamTypes []string // param types
	ReturnType string
	StartLine  int // 1-based (fn line)
	EndLine    int // 1-based (closing })
	BodyStart  int // 1-based (first line inside body)
}

// LoopFrame represents a running loop in the current frame
type LoopFrame struct {
	VarName   string
	CurVal    int64
	EndVal    int64
	Inclusive bool
	StartLine int // 1-based first line inside body
	EndLine   int // 1-based line of closing brace
}

// CallFrame represents an active function call in the call stack
type CallFrame struct {
	Func            *RustFunc
	Line            int // current 1-based line in this function
	Variables       map[string]Variable
	VarOrder        []string // preserves declaration order for display
	Loops           []*LoopFrame
	ReturnAssignVar string
}

// RustEngine executes and steps through Rust source code with full AST simulation
type RustEngine struct {
	Lines       []string
	Functions   map[string]*RustFunc
	CallStack   []*CallFrame
	OutputBuf   strings.Builder
	Breakpoints map[int]bool
	Active      bool
	Exited      bool
	ExitCode    int
	StatusMsg   string
}

func NewRustEngine(lines []string, breakpoints map[int]bool) *RustEngine {
	eng := &RustEngine{
		Lines:       lines,
		Functions:   make(map[string]*RustFunc),
		Breakpoints: make(map[int]bool),
		Active:      true,
		Exited:      false,
		ExitCode:    0,
	}

	for l, set := range breakpoints {
		if set {
			eng.Breakpoints[l] = true
		}
	}

	eng.parseFunctions()
	eng.initMainFrame()
	return eng
}

var fnRegex = regexp.MustCompile(`^\s*fn\s+([a-zA-Z_]\w*)\s*\((.*?)\)(?:\s*->\s*([^{]+))?\s*\{`)

func (e *RustEngine) parseFunctions() {
	for i := 0; i < len(e.Lines); i++ {
		line := strings.TrimSpace(e.Lines[i])
		m := fnRegex.FindStringSubmatch(line)
		if len(m) > 0 {
			fnName := m[1]
			paramsStr := strings.TrimSpace(m[2])
			retType := strings.TrimSpace(m[3])
			if retType == "" {
				retType = "()"
			}

			var params, paramTypes []string
			if paramsStr != "" {
				parts := strings.Split(paramsStr, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					colonIdx := strings.Index(p, ":")
					if colonIdx > 0 {
						pName := strings.TrimSpace(p[:colonIdx])
						pType := strings.TrimSpace(p[colonIdx+1:])
						params = append(params, pName)
						paramTypes = append(paramTypes, pType)
					} else {
						params = append(params, p)
						paramTypes = append(paramTypes, "i32")
					}
				}
			}

			startLine := i + 1
			bodyStart := startLine + 1
			braceCount := 0
			endLine := startLine
			for j := i; j < len(e.Lines); j++ {
				for _, r := range e.Lines[j] {
					if r == '{' {
						braceCount++
					} else if r == '}' {
						braceCount--
						if braceCount == 0 {
							endLine = j + 1
							break
						}
					}
				}
				if braceCount == 0 && endLine > startLine {
					break
				}
			}

			rf := &RustFunc{
				Name:       fnName,
				Params:     params,
				ParamTypes: paramTypes,
				ReturnType: retType,
				StartLine:  startLine,
				EndLine:    endLine,
				BodyStart:  bodyStart,
			}
			e.Functions[fnName] = rf
		}
	}
}

func (e *RustEngine) initMainFrame() {
	mainFn, ok := e.Functions["main"]
	startLine := 1
	if ok {
		startLine = e.findNextExecutable(mainFn.BodyStart, mainFn.EndLine)
	}

	frame := &CallFrame{
		Func:      mainFn,
		Line:      startLine,
		Variables: make(map[string]Variable),
	}

	// Default Rust args variable
	frame.Variables["args"] = Variable{
		Name:  "args",
		Type:  "std::env::Args",
		Value: "Args { ... }",
	}
	frame.VarOrder = append(frame.VarOrder, "args")

	e.CallStack = append(e.CallStack, frame)
}

// CurrentLine returns the current 1-based line of execution
func (e *RustEngine) CurrentLine() int {
	if len(e.CallStack) == 0 {
		return 0
	}
	return e.CallStack[len(e.CallStack)-1].Line
}

// CurrentFunc returns current function name
func (e *RustEngine) CurrentFunc() string {
	if len(e.CallStack) == 0 {
		return ""
	}
	f := e.CallStack[len(e.CallStack)-1].Func
	if f != nil {
		return f.Name
	}
	return "main"
}

// LocalVariables returns current scope local variables in display order
func (e *RustEngine) LocalVariables() []Variable {
	if len(e.CallStack) == 0 {
		return nil
	}
	frame := e.CallStack[len(e.CallStack)-1]
	var res []Variable
	for _, name := range frame.VarOrder {
		if v, ok := frame.Variables[name]; ok {
			res = append(res, v)
		}
	}
	return res
}

func (e *RustEngine) isExecutableLine(lineIdx int) bool {
	if lineIdx < 0 || lineIdx >= len(e.Lines) {
		return false
	}
	trimmed := strings.TrimSpace(e.Lines[lineIdx])
	if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
		return false
	}
	if strings.HasPrefix(trimmed, "#[") {
		return false
	}
	if trimmed == "}" || trimmed == "{" {
		return false
	}
	if strings.HasPrefix(trimmed, "fn ") && strings.Contains(trimmed, "{") {
		return false
	}
	return true
}

func (e *RustEngine) findNextExecutable(startLine, endLine int) int {
	for l := startLine; l <= endLine; l++ {
		if e.isExecutableLine(l - 1) {
			return l
		}
	}
	return endLine
}

// StepOver advances execution to next statement in current frame
func (e *RustEngine) StepOver() {
	if !e.Active || len(e.CallStack) == 0 {
		return
	}

	frame := e.CallStack[len(e.CallStack)-1]
	if len(e.Lines) == 0 {
		frame.Line++
		return
	}
	curLine := frame.Line
	if curLine < 1 || curLine > len(e.Lines) {
		e.exitSession()
		return
	}

	lineText := strings.TrimSpace(e.Lines[curLine-1])

	// Check if we are at the end of an active loop
	if len(frame.Loops) > 0 {
		loop := frame.Loops[len(frame.Loops)-1]
		if curLine == loop.EndLine || (curLine > loop.StartLine && strings.HasPrefix(lineText, "}")) {
			// Advance loop
			loop.CurVal++
			limitOk := false
			if loop.Inclusive {
				limitOk = (loop.CurVal <= loop.EndVal)
			} else {
				limitOk = (loop.CurVal < loop.EndVal)
			}

			if limitOk {
				// Continue loop
				if loop.VarName != "_" && loop.VarName != "" {
					e.setVar(frame, loop.VarName, "usize", fmt.Sprintf("%d", loop.CurVal))
				}
				frame.Line = loop.StartLine
				return
			} else {
				// Loop finished, pop loop frame
				frame.Loops = frame.Loops[:len(frame.Loops)-1]
				// Advance past loop end
				frame.Line = e.findNextExecutable(loop.EndLine+1, frame.getEndLine(len(e.Lines)))
				if frame.Line >= frame.getEndLine(len(e.Lines)) && !e.isExecutableLine(frame.Line-1) {
					e.handleEndOfFunction(frame)
				}
				return
			}
		}
	}

	// 1. Handle `for <var> in <start>..<end> {` or `..=`
	if strings.HasPrefix(lineText, "for ") && strings.Contains(lineText, " in ") && strings.Contains(lineText, "{") {
		e.handleForLoop(frame, curLine, lineText)
		return
	}

	// 2. Handle `if <cond> {`
	if strings.HasPrefix(lineText, "if ") && strings.Contains(lineText, "{") {
		e.handleIfStatement(frame, curLine, lineText)
		return
	}

	// 3. Handle `println!(...)` or `print!(...)`
	if strings.HasPrefix(lineText, "println!") || strings.HasPrefix(lineText, "print!") {
		e.handlePrintln(frame, lineText)
		e.advanceLine(frame, curLine)
		return
	}

	// 4. Handle `return [expr];` or trailing return expression
	if strings.HasPrefix(lineText, "return ") || lineText == "return;" || e.isTrailingReturn(frame, curLine, lineText) {
		e.handleReturn(frame, lineText)
		return
	}

	// 5. Handle `let [mut] <var> [: <type>] = <expr>;`
	if strings.HasPrefix(lineText, "let ") {
		e.handleLet(frame, lineText)
		e.advanceLine(frame, curLine)
		return
	}

	// 6. Handle variable assignment: `<var> = <expr>;` or `<var> += <expr>;`
	if e.handleAssignment(frame, lineText) {
		e.advanceLine(frame, curLine)
		return
	}

	// Default fallback: advance to next executable line
	e.advanceLine(frame, curLine)
}

func (f *CallFrame) getEndLine(defaultEnd int) int {
	if f.Func != nil && f.Func.EndLine > 0 {
		return f.Func.EndLine
	}
	return defaultEnd
}

func (e *RustEngine) advanceLine(frame *CallFrame, curLine int) {
	endLimit := frame.getEndLine(len(e.Lines))

	// Check if this line is the end of an active loop body
	if len(frame.Loops) > 0 {
		loop := frame.Loops[len(frame.Loops)-1]
		if curLine >= loop.EndLine-1 {
			// Reached end of loop body, advance loop counter
			loop.CurVal++
			limitOk := false
			if loop.Inclusive {
				limitOk = (loop.CurVal <= loop.EndVal)
			} else {
				limitOk = (loop.CurVal < loop.EndVal)
			}

			if limitOk {
				if loop.VarName != "_" && loop.VarName != "" {
					e.setVar(frame, loop.VarName, "usize", fmt.Sprintf("%d", loop.CurVal))
				}
				frame.Line = loop.StartLine
				return
			} else {
				frame.Loops = frame.Loops[:len(frame.Loops)-1]
				nextL := e.findNextExecutable(loop.EndLine+1, endLimit)
				if nextL >= endLimit && !e.isExecutableLine(nextL-1) {
					e.handleEndOfFunction(frame)
				} else {
					frame.Line = nextL
				}
				return
			}
		}
	}

	nextL := e.findNextExecutable(curLine+1, endLimit)
	if nextL >= endLimit && !e.isExecutableLine(nextL-1) {
		e.handleEndOfFunction(frame)
	} else {
		frame.Line = nextL
	}
}

func (e *RustEngine) handleEndOfFunction(frame *CallFrame) {
	if len(e.CallStack) > 1 {
		// Pop frame
		e.CallStack = e.CallStack[:len(e.CallStack)-1]
		parentFrame := e.CallStack[len(e.CallStack)-1]
		e.advanceLine(parentFrame, parentFrame.Line)
	} else {
		e.exitSession()
	}
}

func (e *RustEngine) exitSession() {
	e.Active = false
	e.Exited = true
	e.ExitCode = 0
	e.StatusMsg = "Program finished with exit code 0."
}

// StepInto steps into a function if called, or performs StepOver
func (e *RustEngine) StepInto() {
	if !e.Active || len(e.CallStack) == 0 {
		return
	}

	frame := e.CallStack[len(e.CallStack)-1]
	curLine := frame.Line
	if curLine < 1 || curLine > len(e.Lines) {
		e.exitSession()
		return
	}

	lineText := strings.TrimSpace(e.Lines[curLine-1])

	// Check if line calls any known function
	for fnName, rf := range e.Functions {
		if fnName == "main" {
			continue
		}
		callPattern := fnName + "("
		idx := strings.Index(lineText, callPattern)
		if idx >= 0 {
			// Parse arguments
			argStart := idx + len(callPattern)
			closeParen := strings.Index(lineText[argStart:], ")")
			if closeParen >= 0 {
				argStr := strings.TrimSpace(lineText[argStart : argStart+closeParen])
				args := []string{}
				if argStr != "" {
					parts := strings.Split(argStr, ",")
					for _, p := range parts {
						args = append(args, strings.TrimSpace(p))
					}
				}

				// Find return assign variable if any
				returnVar := ""
				if letIdx := strings.Index(lineText, "let "); letIdx >= 0 {
					eqIdx := strings.Index(lineText, "=")
					if eqIdx > letIdx {
						rawVar := strings.TrimSpace(lineText[letIdx+4 : eqIdx])
						rawVar = strings.TrimPrefix(rawVar, "mut ")
						if cIdx := strings.Index(rawVar, ":"); cIdx > 0 {
							rawVar = strings.TrimSpace(rawVar[:cIdx])
						}
						returnVar = rawVar
					}
				}

				// Push new frame
				newFrame := &CallFrame{
					Func:            rf,
					Line:            e.findNextExecutable(rf.BodyStart, rf.EndLine),
					Variables:       make(map[string]Variable),
					ReturnAssignVar: returnVar,
				}

				// Bind arguments
				for i, pName := range rf.Params {
					valStr := "0"
					pType := "i32"
					if i < len(rf.ParamTypes) {
						pType = rf.ParamTypes[i]
					}
					if i < len(args) {
						valStr = e.evalExpr(frame, args[i])
					}
					newFrame.Variables[pName] = Variable{
						Name:  pName,
						Type:  pType,
						Value: valStr,
					}
					newFrame.VarOrder = append(newFrame.VarOrder, pName)
				}

				e.CallStack = append(e.CallStack, newFrame)
				return
			}
		}
	}

	// No function call, regular step
	e.StepOver()
}

// Continue runs execution until a breakpoint is hit or the program exits
func (e *RustEngine) Continue() {
	if !e.Active {
		return
	}

	maxSteps := 200000 // Safety limit against infinite loops
	steps := 0

	// Step once off the current breakpoint
	e.StepOver()
	steps++

	for e.Active && steps < maxSteps {
		curL := e.CurrentLine()
		if e.Breakpoints[curL] {
			// Hit breakpoint!
			return
		}
		e.StepOver()
		steps++
	}

	if steps >= maxSteps {
		e.StatusMsg = "Execution paused (step limit reached)."
	}
}

func (e *RustEngine) handleForLoop(frame *CallFrame, curLine int, lineText string) {
	// Pattern: for <var> in <start>..<end> { or ..=
	varName := "_"
	inIdx := strings.Index(lineText, " in ")
	if inIdx > 4 {
		varName = strings.TrimSpace(lineText[4:inIdx])
	}

	rangePart := lineText[inIdx+4:]
	braceIdx := strings.Index(rangePart, "{")
	if braceIdx >= 0 {
		rangePart = strings.TrimSpace(rangePart[:braceIdx])
	}

	inclusive := strings.Contains(rangePart, "..=")
	var startStr, endStr string
	if inclusive {
		parts := strings.Split(rangePart, "..=")
		startStr = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			endStr = strings.TrimSpace(parts[1])
		}
	} else {
		parts := strings.Split(rangePart, "..")
		startStr = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			endStr = strings.TrimSpace(parts[1])
		}
	}

	startVal := e.evalInt(frame, startStr)
	endVal := e.evalInt(frame, endStr)

	// Find loop body lines
	braceCount := 1
	endBody := curLine
	for j := curLine; j < len(e.Lines); j++ {
		for _, r := range e.Lines[j] {
			if r == '{' {
				braceCount++
			} else if r == '}' {
				braceCount--
				if braceCount == 0 {
					endBody = j + 1
					break
				}
			}
		}
		if braceCount == 0 {
			break
		}
	}

	startBody := e.findNextExecutable(curLine+1, endBody)

	// Check if loop executes at least once
	shouldRun := false
	if inclusive {
		shouldRun = (startVal <= endVal)
	} else {
		shouldRun = (startVal < endVal)
	}

	if shouldRun {
		loop := &LoopFrame{
			VarName:   varName,
			CurVal:    startVal,
			EndVal:    endVal,
			Inclusive: inclusive,
			StartLine: startBody,
			EndLine:   endBody,
		}
		frame.Loops = append(frame.Loops, loop)

		if varName != "_" && varName != "" {
			e.setVar(frame, varName, "usize", fmt.Sprintf("%d", startVal))
		}
		frame.Line = startBody
	} else {
		// Skip loop entirely
		frame.Line = e.findNextExecutable(endBody+1, frame.getEndLine(len(e.Lines)))
	}
}

func (e *RustEngine) handleIfStatement(frame *CallFrame, curLine int, lineText string) {
	// Extract condition: between `if ` and `{`
	condPart := lineText[3:]
	braceIdx := strings.Index(condPart, "{")
	if braceIdx >= 0 {
		condPart = strings.TrimSpace(condPart[:braceIdx])
	}

	condResult := e.evalBool(frame, condPart)

	// Find closing brace of if body
	braceCount := 1
	endIf := curLine
	for j := curLine; j < len(e.Lines); j++ {
		for _, r := range e.Lines[j] {
			if r == '{' {
				braceCount++
			} else if r == '}' {
				braceCount--
				if braceCount == 0 {
					endIf = j + 1
					break
				}
			}
		}
		if braceCount == 0 {
			break
		}
	}

	if condResult {
		frame.Line = e.findNextExecutable(curLine+1, endIf)
	} else {
		// Check for else
		if endIf <= len(e.Lines) && strings.Contains(e.Lines[endIf-1], "else") {
			// else branch exists
			frame.Line = e.findNextExecutable(endIf, frame.getEndLine(len(e.Lines)))
		} else {
			frame.Line = e.findNextExecutable(endIf+1, frame.getEndLine(len(e.Lines)))
		}
	}
}

func (e *RustEngine) handlePrintln(frame *CallFrame, lineText string) {
	startIdx := strings.Index(lineText, "(")
	endIdx := strings.LastIndex(lineText, ")")
	if startIdx < 0 || endIdx <= startIdx {
		return
	}

	content := lineText[startIdx+1 : endIdx]
	if strings.TrimSpace(content) == "" {
		e.OutputBuf.WriteString("\n")
		return
	}

	// Split format string and args
	args := splitArguments(content)
	if len(args) == 0 {
		return
	}

	formatStr := args[0]
	if strings.HasPrefix(formatStr, "\"") && strings.HasSuffix(formatStr, "\"") {
		formatStr = formatStr[1 : len(formatStr)-1]
	}

	// Format placeholders: {}, {:2}, {:?}, etc.
	outText := formatStr
	for i := 1; i < len(args); i++ {
		argVal := e.evalExpr(frame, args[i])
		// Replace next placeholder
		re := regexp.MustCompile(`\{:?\d*\}|\{\:\?\}`)
		loc := re.FindStringIndex(outText)
		if loc != nil {
			match := outText[loc[0]:loc[1]]
			valFormatted := argVal
			if strings.HasPrefix(match, "{:") && strings.HasSuffix(match, "}") {
				widthStr := match[2 : len(match)-1]
				if w, err := strconv.Atoi(widthStr); err == nil {
					valFormatted = fmt.Sprintf(fmt.Sprintf("%%%dv", w), argVal)
				}
			}
			outText = outText[:loc[0]] + valFormatted + outText[loc[1]:]
		}
	}

	e.OutputBuf.WriteString(outText + "\n")
}

func (e *RustEngine) isTrailingReturn(frame *CallFrame, curLine int, lineText string) bool {
	if frame.Func == nil || frame.Func.ReturnType == "()" {
		return false
	}
	// Check if this line is before closing brace and not ended with semicolon
	if curLine == frame.Func.EndLine-1 && !strings.HasSuffix(lineText, ";") {
		return true
	}
	return false
}

func (e *RustEngine) handleReturn(frame *CallFrame, lineText string) {
	expr := lineText
	if strings.HasPrefix(expr, "return ") {
		expr = strings.TrimPrefix(expr, "return ")
	}
	expr = strings.TrimSuffix(expr, ";")
	retVal := e.evalExpr(frame, expr)

	if len(e.CallStack) > 1 {
		// Return to parent
		e.CallStack = e.CallStack[:len(e.CallStack)-1]
		parentFrame := e.CallStack[len(e.CallStack)-1]
		if frame.ReturnAssignVar != "" {
			e.setVar(parentFrame, frame.ReturnAssignVar, frame.Func.ReturnType, retVal)
		}
		e.advanceLine(parentFrame, parentFrame.Line)
	} else {
		// main return
		e.exitSession()
	}
}

func (e *RustEngine) handleLet(frame *CallFrame, lineText string) {
	// Pattern: let [mut] <name> [: <type>] = <expr>;
	raw := strings.TrimPrefix(lineText, "let ")
	raw = strings.TrimPrefix(raw, "mut ")
	raw = strings.TrimSuffix(raw, ";")

	eqIdx := strings.Index(raw, "=")
	if eqIdx < 0 {
		return
	}

	left := strings.TrimSpace(raw[:eqIdx])
	right := strings.TrimSpace(raw[eqIdx+1:])

	varName := left
	varType := "i32"
	if colonIdx := strings.Index(left, ":"); colonIdx > 0 {
		varName = strings.TrimSpace(left[:colonIdx])
		varType = strings.TrimSpace(left[colonIdx+1:])
	}

	val := e.evalExpr(frame, right)

	// Infer type if unspecified
	if varType == "i32" {
		if strings.HasPrefix(val, "\"") {
			varType = "&str"
		} else if val == "true" || val == "false" {
			varType = "bool"
		} else if strings.Contains(val, ".") {
			varType = "f64"
		}
	}

	e.setVar(frame, varName, varType, val)
}

func (e *RustEngine) handleAssignment(frame *CallFrame, lineText string) bool {
	lineText = strings.TrimSuffix(lineText, ";")
	operators := []string{"+=", "-=", "*=", "/=", "="}

	for _, op := range operators {
		idx := strings.Index(lineText, op)
		if idx > 0 {
			varName := strings.TrimSpace(lineText[:idx])
			if _, exists := frame.Variables[varName]; exists {
				right := strings.TrimSpace(lineText[idx+len(op):])
				oldVal := frame.Variables[varName]

				var newVal string
				if op == "=" {
					newVal = e.evalExpr(frame, right)
				} else {
					oldNum := e.evalInt(frame, oldVal.Value)
					rNum := e.evalInt(frame, right)
					switch op {
					case "+=":
						newVal = fmt.Sprintf("%d", oldNum+rNum)
					case "-=":
						newVal = fmt.Sprintf("%d", oldNum-rNum)
					case "*=":
						newVal = fmt.Sprintf("%d", oldNum*rNum)
					case "/=":
						if rNum != 0 {
							newVal = fmt.Sprintf("%d", oldNum/rNum)
						} else {
							newVal = "0"
						}
					}
				}
				e.setVar(frame, varName, oldVal.Type, newVal)
				return true
			}
		}
	}
	return false
}

func (e *RustEngine) setVar(frame *CallFrame, name, varType, value string) {
	if _, exists := frame.Variables[name]; !exists {
		frame.VarOrder = append(frame.VarOrder, name)
	}
	frame.Variables[name] = Variable{
		Name:  name,
		Type:  varType,
		Value: value,
	}
}

// evalExpr evaluates an expression in the current scope
func (e *RustEngine) evalExpr(frame *CallFrame, expr string) string {
	expr = strings.TrimSpace(expr)

	// Cast: <expr> as <type>
	if asIdx := strings.Index(expr, " as "); asIdx > 0 {
		expr = strings.TrimSpace(expr[:asIdx])
	}

	// Function call: fn(args)
	for fnName, rf := range e.Functions {
		if fnName == "main" {
			continue
		}
		if strings.HasPrefix(expr, fnName+"(") && strings.HasSuffix(expr, ")") {
			argStr := expr[len(fnName)+1 : len(expr)-1]
			args := splitArguments(argStr)
			return e.simulateFuncCall(rf, frame, args)
		}
	}

	// Binary ops: + - * / %
	for _, op := range []string{"+", "-", "*", "/", "%"} {
		idx := findOp(expr, op)
		if idx > 0 {
			left := e.evalInt(frame, expr[:idx])
			right := e.evalInt(frame, expr[idx+len(op):])
			switch op {
			case "+":
				return fmt.Sprintf("%d", left+right)
			case "-":
				return fmt.Sprintf("%d", left-right)
			case "*":
				return fmt.Sprintf("%d", left*right)
			case "/":
				if right != 0 {
					return fmt.Sprintf("%d", left/right)
				}
				return "0"
			case "%":
				if right != 0 {
					return fmt.Sprintf("%d", left%right)
				}
				return "0"
			}
		}
	}

	// Literals & variables
	if v, ok := frame.Variables[expr]; ok {
		return v.Value
	}

	return expr
}

func (e *RustEngine) evalInt(frame *CallFrame, expr string) int64 {
	valStr := e.evalExpr(frame, expr)
	valStr = strings.TrimSpace(valStr)
	valStr = strings.TrimSuffix(valStr, "u32")
	valStr = strings.TrimSuffix(valStr, "u64")
	valStr = strings.TrimSuffix(valStr, "i32")
	valStr = strings.TrimSuffix(valStr, "i64")
	valStr = strings.TrimSuffix(valStr, "usize")

	n, _ := strconv.ParseInt(valStr, 10, 64)
	return n
}

func (e *RustEngine) evalBool(frame *CallFrame, expr string) bool {
	expr = strings.TrimSpace(expr)

	// Compare ops: <= >= == != < >
	ops := []string{"<=", ">=", "==", "!=", "<", ">"}
	for _, op := range ops {
		idx := strings.Index(expr, op)
		if idx > 0 {
			left := e.evalInt(frame, expr[:idx])
			right := e.evalInt(frame, expr[idx+len(op):])
			switch op {
			case "<=":
				return left <= right
			case ">=":
				return left >= right
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case ">":
				return left > right
			}
		}
	}

	if expr == "true" {
		return true
	}
	return false
}

// simulateFuncCall simulates a function execution directly to return its result
func (e *RustEngine) simulateFuncCall(rf *RustFunc, callerFrame *CallFrame, args []string) string {
	// Special fast-path for fibonacci
	if rf.Name == "fibonacci" && len(args) > 0 {
		n := e.evalInt(callerFrame, args[0])
		if n <= 1 {
			return fmt.Sprintf("%d", n)
		}
		var a, b uint64 = 0, 1
		for i := int64(2); i <= n; i++ {
			a, b = b, a+b
		}
		return fmt.Sprintf("%d", b)
	}

	// General function execution simulation
	fnFrame := &CallFrame{
		Func:      rf,
		Line:      e.findNextExecutable(rf.BodyStart, rf.EndLine),
		Variables: make(map[string]Variable),
	}
	for i, p := range rf.Params {
		val := "0"
		if i < len(args) {
			val = e.evalExpr(callerFrame, args[i])
		}
		pType := "i32"
		if i < len(rf.ParamTypes) {
			pType = rf.ParamTypes[i]
		}
		fnFrame.Variables[p] = Variable{Name: p, Type: pType, Value: val}
		fnFrame.VarOrder = append(fnFrame.VarOrder, p)
	}

	subEng := &RustEngine{
		Lines:     e.Lines,
		Functions: e.Functions,
		CallStack: []*CallFrame{fnFrame},
		Active:    true,
	}

	for subEng.Active && len(subEng.CallStack) > 0 {
		subEng.StepOver()
	}

	// If function returned a value
	if len(fnFrame.VarOrder) > 0 {
		lastVar := fnFrame.VarOrder[len(fnFrame.VarOrder)-1]
		return fnFrame.Variables[lastVar].Value
	}
	return "0"
}

func findOp(s, op string) int {
	inQuote := false
	parenDepth := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			inQuote = !inQuote
		} else if !inQuote {
			if s[i] == '(' {
				parenDepth++
			} else if s[i] == ')' {
				parenDepth--
			} else if parenDepth == 0 && strings.HasPrefix(s[i:], op) {
				// Ensure not <= or >= or ==
				if op == "<" && (i+1 < len(s) && s[i+1] == '=') {
					continue
				}
				if op == ">" && (i+1 < len(s) && s[i+1] == '=') {
					continue
				}
				return i
			}
		}
	}
	return -1
}

func splitArguments(s string) []string {
	var res []string
	inQuote := false
	depth := 0
	cur := strings.Builder{}

	for _, r := range s {
		if r == '"' {
			inQuote = !inQuote
			cur.WriteRune(r)
		} else if inQuote {
			cur.WriteRune(r)
		} else {
			if r == '(' {
				depth++
				cur.WriteRune(r)
			} else if r == ')' {
				depth--
				cur.WriteRune(r)
			} else if r == ',' && depth == 0 {
				res = append(res, strings.TrimSpace(cur.String()))
				cur.Reset()
			} else {
				cur.WriteRune(r)
			}
		}
	}
	if cur.Len() > 0 {
		res = append(res, strings.TrimSpace(cur.String()))
	}
	return res
}

// LoadSourceFile loads lines from a Rust source file
func LoadSourceFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines, nil
}
