package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"sync"
)

type Position struct {
	X int
	Y int
}

type PositionValue struct {
	Position Position
	Value    uint32
}

type ColorValue struct {
	Value   uint32
	Present bool
}

func submitChoroplethMap(submittedFile multipart.File, data SubmitChoroplethMapData) (*image.RGBA, error) {
	overlayMapImg, err := getOverlayFile()
	if err != nil {
		return nil, err
	}

	overlayBounds := overlayMapImg.Bounds()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(submittedFile); err != nil {
		return nil, fmt.Errorf("oops on read sf")
	}

	submittedImg, err := png.Decode(&buf)
	if err != nil {
		return nil, fmt.Errorf("oops on decode")
	}

	rgbaSubmittedImage := decodeToRGBA(submittedImg)
	submittedImgBounds := rgbaSubmittedImage.Bounds()

	submittedImagePxPerOverlayPx := float64(data.OverlayLocBottomRightY-data.OverlayLocTopLeftY) / float64(overlayBounds.Max.Y)

	colorDataMatrix := initDataMatrix[ColorValue](overlayBounds)

	var wg sync.WaitGroup
	for oy := range overlayBounds.Max.Y {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ox := range overlayBounds.Max.X {
				isRelevant := isRelevant(overlayMapImg, ox, oy)

				sx := data.OverlayLocTopLeftX + int(float64(ox)*submittedImagePxPerOverlayPx)
				sy := data.OverlayLocTopLeftY + int(float64(oy)*submittedImagePxPerOverlayPx)

				newColor := ColorValue{Present: false}

				if sx >= 0 && sx < submittedImgBounds.Max.X && sy >= 0 && sy < submittedImgBounds.Max.Y {
					sr, sg, sb, sa := getRgba(rgbaSubmittedImage, sx, sy)
					scolor := [4]uint8{sr, sg, sb, sa}

					if isRelevant {
						bestLegendItemI := -1
						for i, legendItem := range data.Legend {
							dr, dg, db, da := colorDiff(scolor, legendItem.Color)

							if dr < data.ColorTolerance && dg < data.ColorTolerance && db < data.ColorTolerance && da < data.ColorTolerance {
								if bestLegendItemI != -1 {
									bdr, bdg, bdb, bda := colorDiff(data.Legend[bestLegendItemI].Color, legendItem.Color)
									if dr < bdr && dg < bdg && db < bdb && da < bda {
										bestLegendItemI = i
									}
								} else {
									bestLegendItemI = i
								}
							}
						}

						if bestLegendItemI != -1 {
							value := data.Legend[bestLegendItemI].Value
							if value != nil {
								newColor = ColorValue{Value: uint32(*value * 255), Present: true}
								// newColor = valueColor(*value)
							}
						}
					}
				}

				colorDataMatrix[oy][ox] = newColor

				// newImg.Set(ox, oy, newColor)
			}
		}()
	}

	wg.Wait()

	for range data.BorderTolerance {
		wg = sync.WaitGroup{}
		updates := make(chan PositionValue, overlayBounds.Max.X*overlayBounds.Max.Y)

		islandSizeMatrix := buildIslandMatrix(colorDataMatrix, overlayBounds)

		for y := range overlayBounds.Max.Y {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for x := range overlayBounds.Max.X {
					isRelevant := isRelevant(overlayMapImg, x, y)
					// hasValue := hasValue(colorDataMatrix, x, y)

					if isRelevant {
						colorVal := inheritLargestSiblingColor(colorDataMatrix, islandSizeMatrix, x, y)
						if colorVal.Present {
							updates <- PositionValue{Position: Position{X: x, Y: y}, Value: colorVal.Value}
						}
					}
				}
			}()
		}

		wg.Wait()

		close(updates)

		for update := range updates {
			colorDataMatrix[update.Position.Y][update.Position.X] = ColorValue{Value: update.Value, Present: true}
		}
	}

	newImg := image.NewRGBA(image.Rect(0, 0, overlayBounds.Max.X, overlayBounds.Max.Y))
	for y, row := range colorDataMatrix {
		for x, val := range row {
			if val.Present {
				color := color.RGBA{R: 0, G: uint8(val.Value), B: 0, A: 255}
				newImg.Set(x, y, color)
			}
		}
	}

	return newImg, nil
}

func buildIslandMatrix(colorDataMatrix [][]ColorValue, overlayBounds image.Rectangle) [][]int {
	islandSizeMatrix := initDataMatrix[int](overlayBounds)
	visited := initDataMatrix[bool](overlayBounds)

	for y := range overlayBounds.Max.Y {
		for x := range overlayBounds.Max.X {
			if colorDataMatrix[y][x].Present && !visited[y][x] {
				allIslandPositions := []Position{}
				queue := []Position{{X: x, Y: y}}
				for len(queue) != 0 {
					pos := queue[0]
					queue = queue[1:]

					newPositions := expand(colorDataMatrix, visited, pos.X, pos.Y)

					queue = append(queue, newPositions...)
					allIslandPositions = append(allIslandPositions, newPositions...)
					for _, newPos := range newPositions {
						visited[newPos.Y][newPos.X] = true
					}
				}

				islandSize := len(allIslandPositions)
				for _, pos := range allIslandPositions {
					islandSizeMatrix[pos.Y][pos.X] = islandSize
				}
			}
		}
	}

	return islandSizeMatrix
}

func expand(img [][]ColorValue, visited [][]bool, x, y int) []Position {
	islandColor := img[y][x].Value

	expanded := []Position{}
	if y > 0 && img[y-1][x].Present && img[y-1][x].Value == islandColor && !visited[y-1][x] {
		expanded = append(expanded, Position{X: x, Y: y - 1})
	}
	if x > 0 && img[y][x-1].Present && img[y][x-1].Value == islandColor && !visited[y][x-1] {
		expanded = append(expanded, Position{X: x - 1, Y: y})
	}
	if y < len(img)-1 && img[y+1][x].Present && img[y+1][x].Value == islandColor && !visited[y+1][x] {
		expanded = append(expanded, Position{X: x, Y: y + 1})
	}
	if x < len(img[0])-1 && img[y][x+1].Present && img[y][x+1].Value == islandColor && !visited[y][x+1] {
		expanded = append(expanded, Position{X: x + 1, Y: y})
	}

	return expanded
}

func initDataMatrix[T any](bounds image.Rectangle) [][]T {
	rows := make([][]T, bounds.Max.Y)
	for i := range bounds.Max.Y {
		rows[i] = make([]T, bounds.Max.X)
	}

	return rows
}

func hasValue(img [][]ColorValue, x, y int) bool {
	val := img[y][x]
	return val.Present
}

// func inheritSiblingColor(img [][]ColorValue, x, y int) ColorValue {
// 	colors := [4]ColorValue{}

// 	if x > 0 && hasValue(img, x-1, y) {
// 		// return img.At(x-1, y), true
// 		colors[0] = img[y][x-1]
// 	}
// 	if y > 0 && hasValue(img, x, y-1) {
// 		// return img.At(x, y-1), true
// 		colors[1] = img[y-1][x]
// 	}
// 	if x < len(img[0])-1 && hasValue(img, x+1, y) {
// 		// return img.At(x+1, y), true
// 		colors[2] = img[y][x+1]
// 	}
// 	if y < len(img)-1 && hasValue(img, x, y+1) {
// 		// return img.At(x, y+1), true
// 		colors[3] = img[y+1][x]
// 	}

// 	colorCount := make(map[ColorValue]int, 4)
// 	for _, color := range colors {
// 		if color.Present {
// 			colorCount[color] += 1
// 		}
// 	}

// 	var bestColor ColorValue
// 	highestCount := 0
// 	for color, count := range colorCount {
// 		if count > highestCount {
// 			bestColor = color
// 			highestCount = count
// 		}
// 	}

// 	return bestColor
// }

func inheritLargestSiblingColor(img [][]ColorValue, islandSizeMatrix [][]int, x, y int) ColorValue {
	// positions := []Position{{X: x, Y: y}}
	positions := []Position{}

	if hasValue(img, x, y) {
		positions = append(positions, Position{X: x, Y: y})
	}
	if x > 0 && hasValue(img, x-1, y) {
		positions = append(positions, Position{X: x - 1, Y: y})
	}
	if y > 0 && hasValue(img, x, y-1) {
		positions = append(positions, Position{X: x, Y: y - 1})
	}
	if x < len(img[0])-1 && hasValue(img, x+1, y) {
		positions = append(positions, Position{X: x + 1, Y: y})
	}
	if y < len(img)-1 && hasValue(img, x, y+1) {
		positions = append(positions, Position{X: x, Y: y + 1})
	}
	// if x > 0 {
	// 	positions = append(positions, Position{X: x - 1, Y: y})
	// }
	// if y > 0 {
	// 	positions = append(positions, Position{X: x, Y: y - 1})
	// }
	// if x < len(img[0])-1 {
	// 	positions = append(positions, Position{X: x + 1, Y: y})
	// }
	// if y < len(img)-1 {
	// 	positions = append(positions, Position{X: x, Y: y + 1})
	// }

	// colorCount := make(map[ColorValue]int, 4)
	// for _, color := range colors {
	// 	if color.Present {
	// 		colorCount[color] += 1
	// 	}
	// }

	var bestColor ColorValue
	largestIslandSize := 0
	for _, pos := range positions {
		islandSize := islandSizeMatrix[pos.Y][pos.X]
		if islandSize > largestIslandSize {
			bestColor = img[pos.Y][pos.X]
			largestIslandSize = islandSize
		}
	}

	return bestColor
}
