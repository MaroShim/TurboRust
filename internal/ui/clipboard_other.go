//go:build !windows && !darwin

package ui

import (
	"bytes"
	"os/exec"
)

// SetClipboard copies text to internal clipboard and xclip/wl-copy if available
func SetClipboard(text string) {
	internalClipboard = text
	if err := exec.Command("wl-copy", text).Run(); err != nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = bytes.NewBufferString(text)
		_ = cmd.Run()
	}
}

// GetClipboard retrieves text from wl-paste/xclip or fallback to internal
func GetClipboard() string {
	if out, err := exec.Command("wl-paste").Output(); err == nil && len(out) > 0 {
		return string(out)
	}
	if out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output(); err == nil && len(out) > 0 {
		return string(out)
	}
	return internalClipboard
}
