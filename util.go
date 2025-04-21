package main

import (
	"image"
	"math"
	"path"
	"strings"
)

func decodeToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba
}

func getRgba(img *image.RGBA, x, y int) (uint8, uint8, uint8, uint8) {
	off := img.PixOffset(x, y)
	return img.Pix[off], img.Pix[off+1], img.Pix[off+2], img.Pix[off+3]
}

func colorDiff(x, y [4]uint8) (int, int, int, int) {
	dr := Abs(int(x[0]) - int(y[0]))
	dg := Abs(int(x[1]) - int(y[1]))
	db := Abs(int(x[2]) - int(y[2]))
	da := Abs(int(x[3]) - int(y[3]))
	return dr, dg, db, da
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

func clampedInverseLerp(min, max, val float64) float64 {
	clampedVal := math.Max(math.Min(val, max), min)
	return (clampedVal - min) / (max - min)
}

func stripExtension(filename string) string {
	ext := path.Ext(filename)
	return strings.Replace(filename, ext, "", 1)
}
