package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type LegendItem struct {
	Color [3]uint8 `json:"color"`
	Value *float32 `json:"value"`
}

type SubmitMapData struct {
	Tag           string       `json:"tag"`
	TopLeftDx     int          `json:"topLeftDx"`
	TopLeftDy     int          `json:"topLeftDy"`
	BottomRightDx int          `json:"bottomRightDx"`
	BottomRightDy int          `json:"bottomRightDy"`
	Legend        []LegendItem `json:"legend"`
}

// type MapData struct {
// 	topLeftDx
// 	topLeftDy
// 	bottomRightDx
// 	bottomRightDy
// 	fileData *multipart.FileHeader `form:"file"`
// }

func main() {
	r := gin.Default()

	r.Use(cors.Default())

	r.Static("/assets", "./assets")
	r.POST("/submit-map", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops")
			return
		}

		fileHeader := form.File["file"][0]
		data := form.Value["data"][0]

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops open file")
			return
		}

		defer file.Close()

		var submitMapData SubmitMapData
		err = json.Unmarshal([]byte(data), &submitMapData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops unmarshal "+err.Error())
			return
		}

		err = submitMap(file, submitMapData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		c.File(fmt.Sprintf("./datasets/%s.png", submitMapData.Tag))
	})

	r.Run()
}

func readImage() ([][]bool, error) {
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

func submitMap(submittedFile multipart.File, data SubmitMapData) error {
	overlayMapFile, err := os.Open("./assets/blackwhite.png")
	if err != nil {
		return err
	}

	defer overlayMapFile.Close()

	overlayMapImg, _, err := image.Decode(overlayMapFile)
	if err != nil {
		// return err
		return fmt.Errorf("Oops on decode ol")
	}

	overlayBounds := overlayMapImg.Bounds()

	newImg := image.NewRGBA(image.Rect(0, 0, overlayBounds.Max.X, overlayBounds.Max.Y))

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(submittedFile); err != nil {
		return fmt.Errorf("Oops on read sf")
	}

	// fmt.Println(len(buf.))

	// reader := bytes.NewReader(&buf)
	submittedImg, err := png.Decode(&buf)
	if err != nil {
		return fmt.Errorf("Oops on decode")
	}

	// tf, err := os.Create("./datasets/test.png")
	// if err != nil {
	// 	// return err
	// 	return fmt.Errorf("Oops on create")
	// }
	// defer tf.Close()
	// err = png.Encode(tf, submittedImg)
	// if err != nil {
	// 	// return err
	// 	return fmt.Errorf("oops on encode")
	// }

	submittedImgBounds := submittedImg.Bounds()
	rgbaSubmittedImage := image.NewRGBA(submittedImgBounds)
	draw.Draw(rgbaSubmittedImage, submittedImgBounds, submittedImg, submittedImgBounds.Min, draw.Src)

	fmt.Println(rgbaSubmittedImage.Bounds().Max.X, rgbaSubmittedImage.Bounds().Max.Y)

	tf, err := os.Create("./datasets/test2.png")
	if err != nil {
		// return err
		return fmt.Errorf("Oops on create")
	}
	defer tf.Close()
	err = png.Encode(tf, rgbaSubmittedImage)
	if err != nil {
		// return err
		return fmt.Errorf("oops on encode")
	}

	for oy := range overlayBounds.Max.Y {
		for ox := range overlayBounds.Max.X {
			r, g, b, a := overlayMapImg.At(ox, oy).RGBA()
			isRelevant := r == 0 && g == 0 && b == 0 && a == 0

			sx := ox + data.TopLeftDx
			sy := oy + data.TopLeftDy

			newColor := image.Transparent.C

			if sx < submittedImgBounds.Max.X && sy < submittedImgBounds.Max.Y {
				pixos := rgbaSubmittedImage.PixOffset(sx, sy)
				sr, sg, sb := rgbaSubmittedImage.Pix[pixos], rgbaSubmittedImage.Pix[pixos+1], rgbaSubmittedImage.Pix[pixos+2]
				scolor := [3]uint8{sr, sg, sb}

				if isRelevant {
					bestLegendItemI := -1
					for i, legendItem := range data.Legend {
						dr, dg, db := colorDiff(scolor, legendItem.Color)

						if dr < 8 && dg < 8 && db < 8 {
							if bestLegendItemI != -1 {
								bdr, bdg, bdb := colorDiff(data.Legend[bestLegendItemI].Color, legendItem.Color)
								if dr < bdr && dg < bdg && db < bdb {
									bestLegendItemI = i
								}
							} else {
								bestLegendItemI = i
							}
						}
					}
					// i := slices.IndexFunc(data.Legend, func(legendKey LegendItem) bool {
					// 	return sr == legendKey.Color[0] && sg == legendKey.Color[1] && sb == legendKey.Color[2]
					// })

					// if i != -1 {
					// 	value := data.Legend[i].Value
					// 	if value != nil {
					// 		newColor = color.RGBA{R: 0, G: uint8(*value * 255), B: 0, A: 255}
					// 	}
					// }

					if bestLegendItemI != -1 {
						value := data.Legend[bestLegendItemI].Value
						if value != nil {
							newColor = color.RGBA{R: 0, G: uint8(*value * 255), B: 0, A: 255}
						}
					}
				}
			}

			newImg.Set(ox, oy, newColor)
		}
	}

	newFile, err := os.Create(fmt.Sprintf("./datasets/%s.png", data.Tag))
	if err != nil {
		// return err
		return fmt.Errorf("Oops on create")
	}

	defer newFile.Close()

	err = png.Encode(newFile, newImg)
	if err != nil {
		// return err
		return fmt.Errorf("oops on encode")
	}

	return nil
}

func colorDiff(x, y [3]uint8) (int, int, int) {
	dr := abs(int(x[0]) - int(y[0]))
	dg := abs(int(x[1]) - int(y[1]))
	db := abs(int(x[2]) - int(y[2]))
	return dr, dg, db
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
