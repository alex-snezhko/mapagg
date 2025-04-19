package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type LegendItem struct {
	Color [4]uint8 `json:"color"`
	Value *float64 `json:"value"`
}

type SubmitMapData struct {
	Tag                    string       `json:"tag"`
	OverlayLocTopLeftX     int          `json:"overlayLocTopLeftX"`
	OverlayLocTopLeftY     int          `json:"overlayLocTopLeftY"`
	OverlayLocBottomRightX int          `json:"overlayLocBottomRightX"`
	OverlayLocBottomRightY int          `json:"overlayLocBottomRightY"`
	ColorTolerance         int          `json:"colorTolerance"`
	Legend                 []LegendItem `json:"legend"`
}

type ConfirmMapData struct {
	Tag string `json:"tag"`
}

type AggregateDataTagInfo struct {
	Tag    string  `json:"tag"`
	Weight float64 `json:"weight"`
}

type AggregateDataRequest struct {
	Tags []AggregateDataTagInfo `json:"tags"`
}

type MapAggregationResponse struct {
	Data [][3]float32 `json:"data"`
}

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

		c.File(fmt.Sprintf("./tmp-datasets/%s.png", submitMapData.Tag))
	})

	r.POST("/confirm-map", func(c *gin.Context) {
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

		var confirmMapData ConfirmMapData
		err = json.Unmarshal([]byte(data), &confirmMapData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops unmarshal "+err.Error())
			return
		}

		err = confirmMap(file, confirmMapData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	r.GET("/tags", func(c *gin.Context) {
		results, err := getTags()
		respond(c, results, err)
	})

	r.POST("/aggregate-data", func(c *gin.Context) {
		body := AggregateDataRequest{}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed bad request input "+err.Error())
			return
		}

		val, err := aggregateData(body)
		respond(c, val, err)
	})

	r.Run()
}

func respond[T any](c *gin.Context, val T, err error) {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": val})
	}
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

	overlayMapImg, err := png.Decode(overlayMapFile)
	if err != nil {
		// return err
		return fmt.Errorf("oops on decode ol")
	}

	overlayBounds := overlayMapImg.Bounds()

	newImg := image.NewRGBA(image.Rect(0, 0, overlayBounds.Max.X, overlayBounds.Max.Y))

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(submittedFile); err != nil {
		return fmt.Errorf("oops on read sf")
	}

	submittedImg, err := png.Decode(&buf)
	if err != nil {
		return fmt.Errorf("oops on decode")
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
							newColor = color.RGBA{R: 0, G: uint8(*value * 255), B: 0, A: 255}
						}
					}
				}
			}

			newImg.Set(ox, oy, newColor)
		}
	}

	newFile, err := os.Create(fmt.Sprintf("./tmp-datasets/%s.png", data.Tag))
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

func confirmMap(submittedFile multipart.File, data ConfirmMapData) error {
	newFile, err := os.Create(fmt.Sprintf("./datasets/%s.png", data.Tag))
	if err != nil {
		return err
	}

	defer newFile.Close()

	if _, err := io.Copy(newFile, submittedFile); err != nil {
		return err
	}

	return nil
}

func getTags() ([]string, error) {
	dirents, err := os.ReadDir("./datasets")
	if err != nil {
		return nil, err
	}

	tags := []string{}
	for _, dirent := range dirents {
		tags = append(tags, stripExtension(dirent.Name()))
	}

	return tags, nil
}

func stripExtension(filename string) string {
	ext := path.Ext(filename)
	return strings.Replace(filename, ext, "", 1)
}

func aggregateData(request AggregateDataRequest) ([][3]float64, error) {
	dirents, err := os.ReadDir("./datasets")
	if err != nil {
		return nil, err
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

	everyN := 5

	var wg sync.WaitGroup
	for _, tag := range validTags {
		fullPath := "./datasets/" + tag.Tag + ".png"

		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := readFile(fullPath, tag.Weight/totalWeight, everyN)
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
		return nil, err
	}

	tlLat, tlLong := 40.936336, -74.285709
	brLat, brLong := 40.476305, -73.655104

	allResults := [][][]float64{}
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	width, height := -1, -1
	for _, result := range allResults {
		if height != -1 && len(result) != height {
			return nil, fmt.Errorf("heights do not match for all images")
		}
		height = len(result)

		for _, row := range result {
			if width != -1 && len(row) != width {
				return nil, fmt.Errorf("widths do not match for all images")
			}
			width = len(row)
		}
	}

	multY := (1.0 / float64(height)) * (brLat - tlLat)
	multX := (1.0 / float64(width)) * (brLong - tlLong)

	agg := [][3]float64{}
	for y := range height {
		for x := range width {
			value := 0.0
			for _, result := range allResults {
				value += result[y][x]
			}

			if value > 0 {
				lat := tlLat + float64(y)*multY
				long := tlLong + float64(x)*multX

				agg = append(agg, [3]float64{lat, long, value})
			}
		}
	}

	return agg, err
}

func readFile(filename string, weight float64, everyN int) ([][]float64, error) {
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

	ySamples := bounds.Max.Y / everyN
	xSamples := bounds.Max.X / everyN

	result := make([][]float64, ySamples)

	var wg sync.WaitGroup
	for iy := range ySamples {
		wg.Add(1)
		go func() {
			defer wg.Done()
			row := []float64{}
			for ix := range xSamples {
				_, g, _, _ := getRgba(rgbaFile, ix*everyN, iy*everyN)

				row = append(row, float64(g)/255*weight)
			}
			result[iy] = row
		}()
	}

	wg.Wait()

	return result, nil
}

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
