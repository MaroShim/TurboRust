//go:build windows

package ui

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	moduser32   = syscall.NewLazyDLL("user32.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procOpenClipboard    = moduser32.NewProc("OpenClipboard")
	procCloseClipboard   = moduser32.NewProc("CloseClipboard")
	procEmptyClipboard   = moduser32.NewProc("EmptyClipboard")
	procGetClipboardData = moduser32.NewProc("GetClipboardData")
	procSetClipboardData = moduser32.NewProc("SetClipboardData")

	procGlobalAlloc  = modkernel32.NewProc("GlobalAlloc")
	procGlobalFree   = modkernel32.NewProc("GlobalFree")
	procGlobalLock   = modkernel32.NewProc("GlobalLock")
	procGlobalUnlock = modkernel32.NewProc("GlobalUnlock")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

func writeWindowsClipboard(text string) bool {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return false
	}
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return false
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	bytesCount := uintptr(len(utf16) * 2)
	hMem, _, _ := procGlobalAlloc.Call(gmemMoveable, bytesCount)
	if hMem == 0 {
		return false
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		procGlobalFree.Call(hMem)
		return false
	}

	dest := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16))
	copy(dest, utf16)

	procGlobalUnlock.Call(hMem)

	r, _, _ = procSetClipboardData.Call(cfUnicodeText, hMem)
	if r == 0 {
		procGlobalFree.Call(hMem)
		return false
	}
	return true
}

func readWindowsClipboard() (string, bool) {
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return "", false
	}
	defer procCloseClipboard.Call()

	hMem, _, _ := procGetClipboardData.Call(cfUnicodeText)
	if hMem == 0 {
		return "", false
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return "", false
	}
	defer procGlobalUnlock.Call(hMem)

	u16ptr := (*uint16)(unsafe.Pointer(ptr))
	var length int
	for *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(u16ptr)) + uintptr(length*2))) != 0 {
		length++
	}
	slice := unsafe.Slice(u16ptr, length)
	return syscall.UTF16ToString(slice), true
}

// SetClipboard copies text to internal clipboard and Windows system clipboard
func SetClipboard(text string) {
	internalClipboard = text
	if !writeWindowsClipboard(text) {
		cmd := exec.Command("clip.exe")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
	}
}

// GetClipboard retrieves text from Windows system clipboard or fallback to internal
func GetClipboard() string {
	if text, ok := readWindowsClipboard(); ok && text != "" {
		return text
	}
	return internalClipboard
}
