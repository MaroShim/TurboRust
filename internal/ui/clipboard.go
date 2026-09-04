package ui

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

var internalClipboard string

// SetClipboard copies text to internal clipboard and system clipboard
func SetClipboard(text string) {
	internalClipboard = text
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = bytes.NewBufferString(text)
		_ = cmd.Run()
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard")
		cmd.Stdin = bytes.NewBufferString(text)
		_ = cmd.Run()
	}
}

// GetClipboard retrieves text from system clipboard or fallback to internal
func GetClipboard() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbpaste")
		out, err := cmd.Output()
		if err == nil && len(out) > 0 {
			return string(out)
		}
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-Clipboard")
		out, err := cmd.Output()
		if err == nil {
			s := strings.TrimRight(string(out), "\r\n")
			if s != "" {
				return s
			}
		}
	}
	return internalClipboard
}
