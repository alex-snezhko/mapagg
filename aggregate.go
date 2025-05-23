package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"sync"
)

func aggregateData(request AggregateDataRequest) (MapAggregationResponse, error) {
	var response MapAggregationResponse

	overlayImg, overlayLatLongBounds, err := getOverlayData()
	if err != nil {
		return response, err
	}

	numPixels := overlayImg.Bounds().Max.Y * overlayImg.Bounds().Max.X
	maxAllowedSamples := 200_000
	numSamples := request.SamplingRate * request.SamplingRate
	if numPixels/numSamples > maxAllowedSamples {
		return response, fmt.Errorf("requested sampling rate too low and would generate %d samples, exceeding the maximum allowed of %d, please specify higher value", numPixels/numSamples, maxAllowedSamples)
	}

	validTags, err := filterTags(request.Tags)
	if err != nil {
		return response, err
	}

	allResults, err := computeAllFileValues(validTags, request.SamplingRate, overlayImg)
	if err != nil {
		return response, err
	}

	width, height := -1, -1
	for _, result := range allResults {
		if height != -1 && len(result.Data) != height {
			return response, fmt.Errorf("heights do not match for all images")
		}
		height = len(result.Data)

		for _, row := range result.Data {
			if width != -1 && len(row) != width {
				return response, fmt.Errorf("widths do not match for all images")
			}
			width = len(row)
		}
	}

	gapX, gapY := getOverlayLatLongGaps(width, height, overlayLatLongBounds)

	aggregateData := []LatLongValue{}
	for y := range height {
		for x := range width {
			value := 0.0
			for _, result := range allResults {
				value += result.Data[y][x]
			}

			if value > 0 {
				lat, long := getLatLong(x, y, gapX, gapY, overlayLatLongBounds)

				aggregateData = append(aggregateData, LatLongValue{lat, long, value})
			}
		}
	}

	return MapAggregationResponse{
		AggregateData:  aggregateData,
		ComponentsData: allResults,
		GapY:           gapY,
		GapX:           gapX,
	}, nil
}

func filterTags(tags []AggregateDataTagInfo) ([]AggregateDataTagInfo, error) {
	dirents, err := os.ReadDir("./database/maps")
	if err != nil {
		return nil, err
	}

	validTags := []AggregateDataTagInfo{}
	for _, dirent := range dirents {
		tag, found := Find(tags, func(tag AggregateDataTagInfo) bool { return tag.Tag == stripExtension(dirent.Name()) })
		if !found {
			continue
		}

		validTags = append(validTags, tag)
	}

	return validTags, nil
}

func computeAllFileValues(validTags []AggregateDataTagInfo, samplingRate int, overlayImg image.Image) ([]TaggedImageData, error) {
	totalWeight := 0.0
	for _, t := range validTags {
		totalWeight += t.Weight
	}

	resultsChan := make(chan TaggedImageData, len(validTags))
	errorsChan := make(chan error, len(validTags))

	var wg sync.WaitGroup
	for _, tagInfo := range validTags {
		fullPath := "./database/maps/" + tagInfo.Tag + ".png"

		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := computeFileValues(fullPath, tagInfo.Weight/totalWeight, tagInfo.IsHighGood, samplingRate, overlayImg)
			if err != nil {
				errorsChan <- err
				return
			}

			resultsChan <- TaggedImageData{Tag: tagInfo.Tag, Data: result}
		}()
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	for err := range errorsChan {
		return nil, err
	}

	allResults := []TaggedImageData{}
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	return allResults, nil
}

func computeFileValues(filename string, weight float64, isHighGood bool, samplingRate int, overlayImg image.Image) ([][]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	pngFile, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	img := decodeToRGBA(pngFile)

	bounds := img.Bounds()

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
				topLeftX, topLeftY := ix*samplingRate, iy*samplingRate

				sumValue := 0
				numRelevant := 0
				for offY := range samplingRate {
					for offX := range samplingRate {
						if isWithinOverlay(overlayImg, topLeftX+offX, topLeftY+offY) {
							_, g, _, _ := getRgba(img, topLeftX+offX, topLeftY+offY)
							sumValue += int(g)
							numRelevant++
						}
					}
				}

				avgValue := 0.0
				if numRelevant > 0 {
					avgValue = float64(sumValue) / float64(numRelevant)
				}
				value := avgValue / 255
				if !isHighGood {
					value = 1 - value
				}
				row = append(row, value*weight)
			}
			result[iy] = row
		}()
	}

	wg.Wait()

	return result, nil
}
