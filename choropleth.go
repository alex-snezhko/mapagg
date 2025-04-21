package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
)

func submitChoroplethMap(submittedFile multipart.File, data SubmitChoroplethMapData) (*image.RGBA, error) {
	overlayMapImg, err := getOverlayFile()
	if err != nil {
		return nil, err
	}

	overlayBounds := overlayMapImg.Bounds()

	newImg := image.NewRGBA(image.Rect(0, 0, overlayBounds.Max.X, overlayBounds.Max.Y))

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

	for oy := range overlayBounds.Max.Y {
		for ox := range overlayBounds.Max.X {
			r, g, b, a := overlayMapImg.At(ox, oy).RGBA()
			isRelevant := r == 0 && g == 0 && b == 0 && a != 0

			sx := data.OverlayLocTopLeftX + int(float64(ox)*submittedImagePxPerOverlayPx)
			sy := data.OverlayLocTopLeftY + int(float64(oy)*submittedImagePxPerOverlayPx)

			newColor := image.Transparent.C

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
							newColor = valueColor(*value)
						}
					}
				}
			}

			newImg.Set(ox, oy, newColor)
		}
	}

	return newImg, nil
}
