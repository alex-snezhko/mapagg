package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"math"
	"mime/multipart"
	"slices"
	"strconv"
	"strings"
	"sync"
)

func submitPointsOfInterest(submittedFile multipart.File, data SubmitCoordinatesData) (*image.RGBA, error) {
	overlayMapImg, overlayLatLongBounds, err := getOverlayData()
	if err != nil {
		return nil, err
	}

	overlayBounds := overlayMapImg.Bounds()

	newImg := image.NewRGBA(image.Rect(0, 0, overlayBounds.Max.X, overlayBounds.Max.Y))

	submittedLatLongs, err := readLatLongCsv(submittedFile, data.LatCol, data.LongCol)
	if err != nil {
		return nil, fmt.Errorf("failed to read latlongs CSV: %w", err)
	}

	gapY := (1.0 / float64(overlayBounds.Max.Y)) * (overlayLatLongBounds.BottomRight.Lat - overlayLatLongBounds.TopLeft.Lat)
	gapX := (1.0 / float64(overlayBounds.Max.X)) * (overlayLatLongBounds.BottomRight.Long - overlayLatLongBounds.TopLeft.Long)

	minThresholdRadiusDeg := data.MinThresholdRadiusMiles / MilesPerLatLongDegree
	maxThresholdRadiusDeg := data.MaxThresholdRadiusMiles / MilesPerLatLongDegree

	var wg sync.WaitGroup
	for y := range overlayBounds.Max.Y {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := range overlayBounds.Max.X {
				r, g, b, a := overlayMapImg.At(x, y).RGBA()
				isRelevant := r == 0 && g == 0 && b == 0 && a != 0

				newColor := image.Transparent.C
				if isRelevant {
					lat := overlayLatLongBounds.TopLeft.Lat + float64(y)*gapY
					long := overlayLatLongBounds.TopLeft.Long + float64(x)*gapX

					minDist := math.MaxFloat64
					for _, submittedLatLong := range submittedLatLongs {
						dLat := math.Abs(lat - submittedLatLong.Lat)
						dLong := math.Abs(long - submittedLatLong.Long)
						dist := math.Sqrt(dLat*dLat + dLong*dLong)

						minDist = math.Min(minDist, dist)
					}

					value := clampedInverseLerp(minThresholdRadiusDeg, maxThresholdRadiusDeg, minDist)

					newColor = valueColor(value)
				}

				newImg.Set(x, y, newColor)
			}
		}()
	}

	wg.Wait()

	return newImg, nil
}

func readLatLongCsv(submittedFile multipart.File, latCol string, longCol string) ([]LatLong, error) {
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
	latI := slices.Index(header, latCol)
	if latI == -1 {
		return nil, fmt.Errorf("lat col unexpectedly not found. Found %s", strings.Join(header, ", "))
	}

	longI := slices.Index(header, longCol)
	if longI == -1 {
		return nil, fmt.Errorf("long col unexpectedly not found")
	}

	result := []LatLong{}
	for _, row := range rows[1:] {
		lat, err := strconv.ParseFloat(row[latI], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse lat %s to float: %w", row[latI], err)
		}

		long, err := strconv.ParseFloat(row[longI], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse long %s to float: %w", row[longI], err)
		}

		result = append(result, LatLong{Lat: lat, Long: long})
	}

	return result, nil
}
