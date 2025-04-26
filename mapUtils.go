package main

import (
	"image"
	"image/color"
)

func valueColor(value float64) color.Color {
	return color.RGBA{R: 0, G: uint8(value * 255), B: 0, A: 255}
}

func isRelevant(img image.Image, x, y int) bool {
	r, g, b, a := img.At(x, y).RGBA()
	return r == 0 && g == 0 && b == 0 && a != 0
}
