// Package icons builds the colored status icons used by the tray and window.
// Pure Go (image stdlib) so it stays platform-neutral; the Windows layer
// converts the image.Image to a walk.Icon. Mirrors _make_image/_COLORS in
// tray/tray_icon.py.
package icons

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
)

//go:embed icon_app.png
var appIconPNG []byte

// AppIcon returns the application's heartbeat icon (used for the window title bar).
func AppIcon() image.Image {
	img, _ := png.Decode(bytes.NewReader(appIconPNG))
	return img
}

// State colors match the badge: green / orange / red.
var palette = map[string]color.RGBA{
	"active":   {39, 174, 96, 255},
	"paused":   {230, 126, 34, 255},
	"inactive": {192, 57, 43, 255},
}

// Circle returns a 64x64 RGBA image: a filled circle in the state color on a
// transparent background. Unknown states fall back to gray.
func Circle(state string) image.Image {
	const size = 64
	fill, ok := palette[state]
	if !ok {
		fill = palette["inactive"]
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Inset matches the Python draw.ellipse([4,4,size-4,size-4]).
	const inset = 4
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - inset

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				img.SetRGBA(x, y, fill)
			}
		}
	}
	return img
}
