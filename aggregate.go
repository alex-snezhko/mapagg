package main

import (
	"fmt"
	"image/png"
	"os"
	"sync"
)

func getTags() ([]string, error) {
	dirents, err := os.ReadDir("./database/maps")
	if err != nil {
		return nil, err
	}

	tags := []string{}
	for _, dirent := range dirents {
		tags = append(tags, stripExtension(dirent.Name()))
	}

	return tags, nil
}

func aggregateData(request AggregateDataRequest) (MapAggregationResponse, error) {
	var response MapAggregationResponse

	overlayFile, err := getOverlayFile()
	if err != nil {
		return response, err
	}

	numPixels := overlayFile.Bounds().Max.Y * overlayFile.Bounds().Max.X
	maxAllowedSamples := 200_000
	numSamples := request.SamplingRate * request.SamplingRate
	if numPixels/numSamples > maxAllowedSamples {
		return response, fmt.Errorf("requested sampling rate too low and would generate %d samples, exceeding the maximum allowed of %d, please specify higher value", numPixels/numSamples, maxAllowedSamples)
	}

	dirents, err := os.ReadDir("./database/maps")
	if err != nil {
		return response, err
	}

	totalWeight := 0.0
	for _, t := range request.Tags {
		totalWeight += t.Weight
	}

	validTags := []AggregateDataTagInfo{}
	for _, dirent := range dirents {
		tag, found := Find(request.Tags, func(tag AggregateDataTagInfo) bool { return tag.Tag == stripExtension(dirent.Name()) })
		if !found {
			continue
		}

		validTags = append(validTags, tag)
	}

	resultsChan := make(chan [][]float64, len(validTags))
	errorsChan := make(chan error, len(validTags))

	var wg sync.WaitGroup
	for _, tag := range validTags {
		fullPath := "./database/maps/" + tag.Tag + ".png"

		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := readFile(fullPath, tag.Weight/totalWeight, request.SamplingRate)
			if err != nil {
				errorsChan <- err
				return
			}

			resultsChan <- result
		}()
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	for err := range errorsChan {
		return response, err
	}

	overlayLatLongBounds, err := getOverlayBounds()
	if err != nil {
		return response, err
	}

	allResults := [][][]float64{}
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	width, height := -1, -1
	for _, result := range allResults {
		if height != -1 && len(result) != height {
			return response, fmt.Errorf("heights do not match for all images")
		}
		height = len(result)

		for _, row := range result {
			if width != -1 && len(row) != width {
				return response, fmt.Errorf("widths do not match for all images")
			}
			width = len(row)
		}
	}

	gapY := (1.0 / float64(height)) * (overlayLatLongBounds.BottomRight.Lat - overlayLatLongBounds.TopLeft.Lat)
	gapX := (1.0 / float64(width)) * (overlayLatLongBounds.BottomRight.Long - overlayLatLongBounds.TopLeft.Long)

	data := [][3]float64{}
	for y := range height {
		for x := range width {
			value := 0.0
			for _, result := range allResults {
				value += result[y][x]
			}

			if value > 0 {
				lat := overlayLatLongBounds.TopLeft.Lat + float64(y)*gapY
				long := overlayLatLongBounds.TopLeft.Long + float64(x)*gapX

				data = append(data, [3]float64{lat, long, value})
			}
		}
	}

	return MapAggregationResponse{Data: data, GapY: gapY, GapX: gapX}, nil
}

func readFile(filename string, weight float64, samplingRate int) ([][]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	pngFile, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	rgbaFile := decodeToRGBA(pngFile)

	bounds := rgbaFile.Bounds()

	ySamples := bounds.Max.Y / samplingRate
	xSamples := bounds.Max.X / samplingRate

	result := make([][]float64, ySamples)

	var wg sync.WaitGroup
	for iy := range ySamples {
		wg.Add(1)
		go func() {
			defer wg.Done()
			row := []float64{}
			for ix := range xSamples {
				_, g, _, _ := getRgba(rgbaFile, ix*samplingRate, iy*samplingRate)

				row = append(row, float64(g)/255*weight)
			}
			result[iy] = row
		}()
	}

	wg.Wait()

	return result, nil
}
