//go:build darwin

package ui

import (
	"bytes"
	"os/exec"
)

// SetClipboard copies text to internal clipboard and macOS pbcopy
func SetClipboard(text string) {
	internalClipboard = text
	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewBufferString(text)
	_ = cmd.Run()
}

// GetClipboard retrieves text from macOS pbpaste or fallback to internal
func GetClipboard() string {
	cmd := exec.Command("pbpaste")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return string(out)
	}
	return internalClipboard
}
