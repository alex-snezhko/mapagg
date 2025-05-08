package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type LatLong struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type OverlayBounds struct {
	TopLeft     LatLong `json:"topLeft"`
	BottomRight LatLong `json:"bottomRight"`
}

type LegendItem struct {
	Color [4]uint8 `json:"color"`
	Value *float64 `json:"value"`
}

type SubmitChoroplethMapData struct {
	Tag                    string       `json:"tag"`
	OverlayLocTopLeftX     int          `json:"overlayLocTopLeftX"`
	OverlayLocTopLeftY     int          `json:"overlayLocTopLeftY"`
	OverlayLocBottomRightX int          `json:"overlayLocBottomRightX"`
	OverlayLocBottomRightY int          `json:"overlayLocBottomRightY"`
	ColorTolerance         int          `json:"colorTolerance"`
	BorderTolerance        int          `json:"borderTolerance"`
	Legend                 []LegendItem `json:"legend"`
}

type SubmitFileData struct {
	Data string                `form:"data" binding:"required"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type SubmitPointsOfInterestFromCsvData struct {
	Tag                     string  `json:"tag"`
	MinThresholdRadiusMiles float64 `json:"minThresholdRadiusMiles"`
	MaxThresholdRadiusMiles float64 `json:"maxThresholdRadiusMiles"`
	LatCol                  string  `json:"latCol"`
	LongCol                 string  `json:"longCol"`
	WeightCol               *string `json:"weightCol"`
}

type PointOfInterest struct {
	LatLong LatLong `json:"latLong"`
	Weight  float64 `json:"weight"`
}

type SubmitPointsOfInterestData struct {
	Tag                     string            `json:"tag"`
	PointsOfInterest        []PointOfInterest `json:"pointsOfInterest"`
	MinThresholdRadiusMiles float64           `json:"minThresholdRadiusMiles"`
	MaxThresholdRadiusMiles float64           `json:"maxThresholdRadiusMiles"`
}

type ConfirmMapData struct {
	Tag string `json:"tag"`
}

type AggregateDataTagInfo struct {
	Tag    string  `json:"tag"`
	Weight float64 `json:"weight"`
}

type AggregateDataRequest struct {
	Tags         []AggregateDataTagInfo `json:"tags"`
	SamplingRate int                    `json:"samplingRate"`
}

type LatLongValue = [3]float64

type TaggedImageData struct {
	Tag  string      `json:"tag"`
	Data [][]float64 `json:"data"`
}

type MapAggregationResponse struct {
	AggregateData  []LatLongValue    `json:"aggregateData"`
	ComponentsData []TaggedImageData `json:"componentsData"`
	GapY           float64           `json:"gapY"`
	GapX           float64           `json:"gapX"`
}

const MilesPerLatLongDegree float64 = 69.172

func main() {
	r := gin.Default()

	r.Use(cors.Default())

	r.Static("/assets", "./assets")
	r.POST("/submit-choropleth-map", func(c *gin.Context) {
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

		var submitMapData SubmitChoroplethMapData
		err = json.Unmarshal([]byte(data), &submitMapData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops unmarshal "+err.Error())
			return
		}

		newImg, err := submitChoroplethMap(file, submitMapData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		tmpFilePath, err := writeTmpFile(newImg, submitMapData.Tag)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		c.File(tmpFilePath)
	})

	r.POST("/submit-coordinates-from-csv", func(c *gin.Context) {
		var fileData SubmitFileData

		if err := c.ShouldBind(&fileData); err != nil {
			c.JSON(http.StatusBadRequest, "Oops could not bind")
			return
		}

		file, err := fileData.File.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops open file")
			return
		}

		defer file.Close()

		var submitCoordinatesData SubmitPointsOfInterestFromCsvData
		err = json.Unmarshal([]byte(fileData.Data), &submitCoordinatesData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops unmarshal "+err.Error())
			return
		}

		newImg, err := submitPointsOfInterestFromCsv(file, submitCoordinatesData)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		tmpFilePath, err := writeTmpFile(newImg, submitCoordinatesData.Tag)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		c.File(tmpFilePath)
	})

	r.POST("/submit-coordinates", func(c *gin.Context) {
		var data SubmitPointsOfInterestData

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, "Oops could not bind")
			return
		}

		newImg, err := submitPointsOfInterest(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		tmpFilePath, err := writeTmpFile(newImg, data.Tag)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops failed "+err.Error())
			return
		}

		c.File(tmpFilePath)
	})

	r.POST("/confirm-map", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops")
			return
		}

		fileHeader := form.File["file"][0]
		fileData := form.Value["data"][0]

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, "Oops open file")
			return
		}

		defer file.Close()

		var confirmMapData ConfirmMapData
		err = json.Unmarshal([]byte(fileData), &confirmMapData)
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

	r.GET("/overlay-bounds", func(c *gin.Context) {
		result, err := getOverlayBounds()
		respond(c, result, err)
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

func writeTmpFile(img *image.RGBA, tag string) (string, error) {
	tmpFilepath := fmt.Sprintf("./tmp-database/%s.png", tag)

	newFile, err := os.Create(tmpFilepath)
	if err != nil {
		return "", fmt.Errorf("error writing temp data: %w", err)
	}

	defer newFile.Close()

	err = png.Encode(newFile, img)
	if err != nil {
		return "", fmt.Errorf("error encoding file png: %w", err)
	}

	return tmpFilepath, nil
}

func confirmMap(submittedFile multipart.File, data ConfirmMapData) error {
	newFile, err := os.Create(fmt.Sprintf("./database/maps/%s.png", data.Tag))
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
