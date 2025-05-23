package main

import (
	"image"
	"image/color"
)

func valueColor(value float64) color.Color {
	return color.RGBA{R: 0, G: uint8((1.0 - value) * 255), B: 0, A: 255}
	// return color.RGBA{R: 75, G: uint8(75 + ((1.0 - value) * 180)), B: 75, A: 255}
}

func isWithinOverlay(overlayImg image.Image, x, y int) bool {
	r, g, b, a := overlayImg.At(x, y).RGBA()
	return r == 0 && g == 0 && b == 0 && a != 0
}

func getOverlayLatLongGaps(width, height int, overlayLatLongBounds OverlayBounds) (float64, float64) {
	gapX := (1.0 / float64(width)) * (overlayLatLongBounds.BottomRight.Long - overlayLatLongBounds.TopLeft.Long)
	gapY := (1.0 / float64(height)) * (overlayLatLongBounds.TopLeft.Lat - overlayLatLongBounds.BottomRight.Lat)

	return gapX, gapY
}

func getLatLong(x, y int, gapX, gapY float64, overlayLatLongBounds OverlayBounds) (float64, float64) {
	lat := overlayLatLongBounds.TopLeft.Lat - float64(y)*gapY
	long := overlayLatLongBounds.TopLeft.Long + float64(x)*gapX

	return lat, long
}
