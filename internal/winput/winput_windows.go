//go:build windows

// Package winput simulates mouse/keyboard activity via the Win32 SendInput API,
// a Go port of core/win_input.py. No third-party dependencies — it calls
// user32!SendInput directly, the same API pyautogui used internally.
package winput

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	inputMouse    = 0
	inputKeyboard = 1

	mouseeventfMove = 0x0001
	keyeventfKeyUp  = 0x0002
)

// VKCodes maps the keystroke options exposed in Settings to virtual-key codes.
// Mirror of VK_CODES in core/win_input.py.
var VKCodes = map[string]uint16{
	"shift": 0x10, "ctrl": 0x11, "alt": 0x12,
	"f13": 0x7C, "f14": 0x7D, "f15": 0x7E, "f16": 0x7F,
	"scrolllock": 0x91, "numlock": 0x90, "capslock": 0x14,
}

// INPUT structs laid out to match the Win32 ABI on amd64. Both variants are the
// same size (40 bytes) as the C INPUT union, so SendInput's cb is sizeof of the
// struct passed.
type mouseInput struct {
	typ         uint32
	_           uint32 // pad: union is 8-byte aligned on x64
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type keybdInput struct {
	typ         uint32
	_           uint32
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
	_           [8]byte // pad to the union (mouseInput) size
}

var (
	user32        = windows.NewLazySystemDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

// Available reports whether the native backend can be used (always true on
// Windows).
func Available() bool { return true }

// MoveMouse moves the cursor by (dx, dy) pixels, relative.
func MoveMouse(dx, dy int) {
	in := mouseInput{
		typ:     inputMouse,
		dx:      int32(dx),
		dy:      int32(dy),
		dwFlags: mouseeventfMove,
	}
	procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
}

// TapKey presses and releases the given virtual-key code.
func TapKey(vk uint16) {
	down := keybdInput{typ: inputKeyboard, wVk: vk}
	up := keybdInput{typ: inputKeyboard, wVk: vk, dwFlags: keyeventfKeyUp}
	procSendInput.Call(1, uintptr(unsafe.Pointer(&down)), unsafe.Sizeof(down))
	procSendInput.Call(1, uintptr(unsafe.Pointer(&up)), unsafe.Sizeof(up))
}
