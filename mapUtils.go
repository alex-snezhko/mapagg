package main

import (
	"image"
	"image/color"
)

func valueColor(value float64) color.Color {
	return color.RGBA{R: 0, G: uint8((1.0 - value) * 255), B: 0, A: 255}
	// return color.RGBA{R: 75, G: uint8(75 + ((1.0 - value) * 180)), B: 75, A: 255}
}

func isRelevant(overlayImg image.Image, x, y int) bool {
	r, g, b, a := overlayImg.At(x, y).RGBA()
	return r == 0 && g == 0 && b == 0 && a != 0
}
