package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

type Position struct {
	X int
	Y int
}

type PositionValue struct {
	Position Position
	Value    float64
}

type ColorValue struct {
	Value           float64
	IsWithinOverlay bool
	IsValueFound    bool
}

type LocationValue struct {
	Location string
	Value    float64
}

func submitChoroplethMap(submittedFile io.Reader, data SubmitChoroplethMapData) (*image.RGBA, error) {
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
				isRelevant := isWithinOverlay(overlayMapImg, ox, oy)

				sx := data.OverlayLocTopLeftX + int(float64(ox)*submittedImagePxPerOverlayPx)
				sy := data.OverlayLocTopLeftY + int(float64(oy)*submittedImagePxPerOverlayPx)

				newColor := ColorValue{IsWithinOverlay: false}

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
								newColor = ColorValue{Value: *value, IsWithinOverlay: true}
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
					isRelevant := isWithinOverlay(overlayMapImg, x, y)
					// hasValue := hasValue(colorDataMatrix, x, y)

					if isRelevant {
						colorVal := inheritLargestSiblingColor(colorDataMatrix, islandSizeMatrix, x, y)
						if colorVal.IsWithinOverlay {
							updates <- PositionValue{Position: Position{X: x, Y: y}, Value: colorVal.Value}
						}
					}
				}
			}()
		}

		wg.Wait()

		close(updates)

		for update := range updates {
			colorDataMatrix[update.Position.Y][update.Position.X] = ColorValue{Value: update.Value, IsWithinOverlay: true}
		}
	}

	newImg := colorDataMatrixToImage(colorDataMatrix, overlayBounds)

	return newImg, nil
}

func buildIslandMatrix(colorDataMatrix [][]ColorValue, overlayBounds image.Rectangle) [][]int {
	islandSizeMatrix := initDataMatrix[int](overlayBounds)
	visited := initDataMatrix[bool](overlayBounds)

	for y := range overlayBounds.Max.Y {
		for x := range overlayBounds.Max.X {
			if colorDataMatrix[y][x].IsWithinOverlay && !visited[y][x] {
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
	if y > 0 && img[y-1][x].IsWithinOverlay && img[y-1][x].Value == islandColor && !visited[y-1][x] {
		expanded = append(expanded, Position{X: x, Y: y - 1})
	}
	if x > 0 && img[y][x-1].IsWithinOverlay && img[y][x-1].Value == islandColor && !visited[y][x-1] {
		expanded = append(expanded, Position{X: x - 1, Y: y})
	}
	if y < len(img)-1 && img[y+1][x].IsWithinOverlay && img[y+1][x].Value == islandColor && !visited[y+1][x] {
		expanded = append(expanded, Position{X: x, Y: y + 1})
	}
	if x < len(img[0])-1 && img[y][x+1].IsWithinOverlay && img[y][x+1].Value == islandColor && !visited[y][x+1] {
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
	return val.IsWithinOverlay
}

func submitChoroplethMapFromCsv(geoJsonFile, locationCsvFile io.Reader, data SubmitChoroplethMapFromCsvData) (*image.RGBA, error) {
	geoJsonBytes, err := io.ReadAll(geoJsonFile)
	if err != nil {
		return nil, err
	}

	fc, err := geojson.UnmarshalFeatureCollection(geoJsonBytes)
	if err != nil {
		return nil, err
	}

	for _, feature := range fc.Features {
		featureNameObj, found := feature.Properties[data.GeoJsonNameProperty]
		if !found {
			return nil, fmt.Errorf("could not find name for feature")
		}

		_, isString := featureNameObj.(string)
		if !isString {
			return nil, fmt.Errorf("feature name not string")
		}

		geoJsonType := feature.Geometry.GeoJSONType()
		if geoJsonType != "MultiPolygon" && geoJsonType != "Polygon" {
			return nil, fmt.Errorf("geometry not Polygon nor MultiPolygon")
		}
	}

	locationValues, err := readLocationValuesFromCsv(locationCsvFile, data.CsvNameColumn, data.CsvValueColumn)
	if err != nil {
		return nil, err
	}

	locationValByFeatureName, err := getLocationValByFeatureName(fc, locationValues, data)
	if err != nil {
		return nil, err
	}

	overlayMapImg, overlayLatLongBounds, err := getOverlayData()
	if err != nil {
		return nil, err
	}

	overlayBounds := overlayMapImg.Bounds()

	gapX, gapY := getOverlayLatLongGaps(overlayBounds.Max.X, overlayBounds.Max.Y, overlayLatLongBounds)

	colorDataMatrix := initDataMatrix[ColorValue](overlayBounds)

	var wg sync.WaitGroup
	for y := range overlayBounds.Max.Y {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := range overlayBounds.Max.X {
				isWithinOverlay := isWithinOverlay(overlayMapImg, x, y)

				var newColor ColorValue
				if !isWithinOverlay {
					newColor = ColorValue{IsWithinOverlay: false}
				} else {
					newColor = ColorValue{IsWithinOverlay: true, IsValueFound: false}
					lat, long := getLatLong(x, y, gapX, gapY, overlayLatLongBounds)

					point := orb.Point{long, lat}

					for _, feature := range fc.Features {
						featureName := feature.Properties[data.GeoJsonNameProperty].(string)

						locationVal, found := locationValByFeatureName[featureName]
						if !found {
							continue
						}

						geoJsonType := feature.Geometry.GeoJSONType()

						isWithinBounds := false
						if geoJsonType == "Polygon" {
							poly := feature.Geometry.(orb.Polygon)
							isWithinBounds = planar.PolygonContains(poly, point)
						} else if geoJsonType == "MultiPolygon" {
							polys := feature.Geometry.(orb.MultiPolygon)
							isWithinBounds = planar.MultiPolygonContains(polys, point)
						} else {
							panic("Impossible: geoJson type not Polygon nor MultiPolygon and not asserted previously")
						}

						if !isWithinBounds {
							continue
						}

						value := clampedInverseLerp(data.LowerBoundThreshold, data.UpperBoundThreshold, locationVal.Value)
						newColor = ColorValue{Value: value, IsWithinOverlay: true, IsValueFound: true}
						break
					}
				}

				colorDataMatrix[y][x] = newColor
			}
		}()
	}

	wg.Wait()

	newImg := colorDataMatrixToImage(colorDataMatrix, overlayBounds)

	return newImg, nil
}

func getLocationValByFeatureName(fc *geojson.FeatureCollection, locationValues []LocationValue, data SubmitChoroplethMapFromCsvData) (map[string]LocationValue, error) {
	nonAlphaNumRegex := regexp.MustCompile("[^a-zA-Z0-9 ]+")
	whitespaceRegex := regexp.MustCompile("[ _-]+")

	locationValByFeatureName := make(map[string]LocationValue)

	noMatchFeatureNames := []string{}
	for _, feature := range fc.Features {
		featureName := feature.Properties[data.GeoJsonNameProperty].(string)
		normalizedFeatureName := normalizeName(nonAlphaNumRegex, whitespaceRegex, featureName)

		if data.AllowNameMatchingLeniency {
			validLocations := []LocationValue{}
			for _, locationVal := range locationValues {
				normalizedLocationName := normalizeName(nonAlphaNumRegex, whitespaceRegex, locationVal.Location)
				if strings.Contains(normalizedLocationName, normalizedFeatureName) || strings.Contains(normalizedFeatureName, normalizedLocationName) {
					validLocations = append(validLocations, locationVal)
				}
			}

			if len(validLocations) == 0 {
				noMatchFeatureNames = append(noMatchFeatureNames, featureName)
				continue
			}

			bestMatchDistance := math.MaxInt
			var bestMatch LocationValue
			for _, locationVal := range validLocations {
				normalizedLocationName := normalizeName(nonAlphaNumRegex, whitespaceRegex, locationVal.Location)
				dist := fuzzy.LevenshteinDistance(normalizedLocationName, normalizedFeatureName)
				if dist < bestMatchDistance {
					bestMatchDistance = dist
					bestMatch = locationVal
				}
			}

			locationValByFeatureName[featureName] = bestMatch
		} else {
			locationValI := slices.IndexFunc(locationValues, func(locationVal LocationValue) bool {
				normalizedLocationName := normalizeName(nonAlphaNumRegex, whitespaceRegex, locationVal.Location)
				return normalizedLocationName == normalizedFeatureName
			})

			if locationValI == -1 {
				noMatchFeatureNames = append(noMatchFeatureNames, featureName)
				continue
			}

			locationValByFeatureName[featureName] = locationValues[locationValI]
		}
	}

	if len(noMatchFeatureNames) > 0 && !data.SkipMissing {
		return nil, fmt.Errorf("no matches found for locations %s", strings.Join(noMatchFeatureNames, ", "))
	}

	return locationValByFeatureName, nil
}

func normalizeName(nonAlphaNumRegex, whitespaceRegex *regexp.Regexp, name string) string {
	name = whitespaceRegex.ReplaceAllString(name, " ")
	name = nonAlphaNumRegex.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	return name
}

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

func colorDataMatrixToImage(colorDataMatrix [][]ColorValue, overlayBounds image.Rectangle) *image.RGBA {
	newImg := image.NewRGBA(image.Rect(0, 0, overlayBounds.Max.X, overlayBounds.Max.Y))
	for y, row := range colorDataMatrix {
		for x, val := range row {
			if val.IsWithinOverlay {
				var pixelColor color.Color
				if val.IsValueFound {
					pixelColor = color.RGBA{R: 0, G: uint8(val.Value * 255), B: 0, A: 255}
				} else {
					pixelColor = color.RGBA{R: 220, G: 0, B: 0, A: 255}
				}

				newImg.Set(x, y, pixelColor)
			}
		}
	}

	return newImg
}

func readLocationValuesFromCsv(submittedFile io.Reader, nameCol, valCol string) ([]LocationValue, error) {
	var buf bytes.Buffer
	bytesRead, err := buf.ReadFrom(submittedFile)
	if err != nil {
		return nil, fmt.Errorf("oops on read sf")
	}

	if bytesRead > 10_000_000 {
		return nil, fmt.Errorf("max file size of 10MB exceeded")
	}

	csvReader := csv.NewReader(&buf)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV from file: %w", err)
	}

	if len(rows) < 1 {
		return nil, fmt.Errorf("read csv unexpectedly empty")
	}

	header := rows[0]
	nameI := slices.Index(header, nameCol)
	if nameI == -1 {
		return nil, fmt.Errorf("name col unexpectedly not found. Found %s", strings.Join(header, ", "))
	}

	valI := slices.Index(header, valCol)
	if valI == -1 {
		return nil, fmt.Errorf("value col unexpectedly not found")
	}

	result := []LocationValue{}
	for _, row := range rows[1:] {
		name := row[nameI]

		val, err := strconv.ParseFloat(row[valI], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse long %s to float: %w", row[valI], err)
		}

		result = append(result, LocationValue{
			Location: name,
			Value:    val,
		})
	}

	return result, nil
}
