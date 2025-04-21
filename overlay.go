package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
)

func readOverlay() ([][]bool, error) {
	file, err := os.Open("./nychousingcost.png")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()

	width, height := bounds.Max.X, bounds.Max.Y
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	var pixels [][]bool
	for y := range height {
		var row []bool
		for x := range width {
			r, g, b, a := img.At(x, y).RGBA()
			isLand := a > 240 || (r > 240 && g > 240 && b > 240)

			newColor := image.Transparent.C
			if isLand {
				newColor = image.Black.C
			}

			newImg.Set(x, y, newColor)
			row = append(row, isLand)
		}

		pixels = append(pixels, row)
	}

	newFile, err := os.Create("./assets/blackwhite.png")
	if err != nil {
		return nil, err
	}

	defer newFile.Close()

	err = png.Encode(newFile, newImg)
	if err != nil {
		return nil, err
	}

	return pixels, nil
}

func getOverlayFile() (image.Image, error) {
	overlayMapFile, err := os.Open("./assets/blackwhite.png")
	if err != nil {
		return nil, fmt.Errorf("failed to read overlay image: %w", err)
	}

	defer overlayMapFile.Close()

	overlayMapImg, err := png.Decode(overlayMapFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode overlay map: %w", err)
	}

	return overlayMapImg, nil
}

func getOverlayBounds() (OverlayBounds, error) {
	var overlayData OverlayBounds

	overlayDataBytes, err := os.ReadFile("./database/overlayData.json")
	if err != nil {
		return overlayData, fmt.Errorf("failed to read overlay data: %w", err)
	}

	if err = json.Unmarshal(overlayDataBytes, &overlayData); err != nil {
		return overlayData, fmt.Errorf("failed to parse overlay data json: %w", err)
	}

	return overlayData, nil
}

func getOverlayData() (image image.Image, bounds OverlayBounds, err error) {
	image, err = getOverlayFile()
	if err != nil {
		return
	}

	bounds, err = getOverlayBounds()
	if err != nil {
		return
	}

	return
}
