//go:build !windows

// Non-Windows stub so the core packages build and test on any platform. The app
// itself is Windows-only; on other platforms the engine degrades to "simulado".
package winput

// VKCodes is empty off Windows; no keys resolve.
var VKCodes = map[string]uint16{}

// Available reports false: no native input backend outside Windows.
func Available() bool { return false }

// MoveMouse is a no-op off Windows.
func MoveMouse(dx, dy int) {}

// TapKey is a no-op off Windows.
func TapKey(vk uint16) {}
