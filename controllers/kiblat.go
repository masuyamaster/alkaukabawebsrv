package controllers

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KiblatPost struct {
	Lat  float32 `json:"lat" binding:"required"`
	Long float32 `json:"long" binding:"required"`
}

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

// fungsi konversi radian ke derajat
func rad2deg(r float64) float64 {
	return r * 180.0 / math.Pi
}

func ArahKiblat(lat, lon float64) float64 {
	latKabah := deg2rad(21.4167)
	lonKabah := deg2rad(39.8262)

	// lokasi
	phi := deg2rad(lat)
	lambda := deg2rad(lon)

	// beda bujur (Ka'bah - lokasi)
	dLon := lonKabah - lambda

	y := math.Sin(dLon)
	x := math.Cos(phi)*math.Tan(latKabah) - math.Sin(phi)*math.Cos(dLon)
	println("y ", y, "x", x)

	theta := math.Atan2(y, x) // langsung kuadran benar
	azimuth := rad2deg(theta)
	azimuth2 := rad2deg(x)
	azimuth3 := rad2deg(y)
	// fmt.Printf("azimuth %.2f %.2f %.2f %.2f", azimuth, theta, azimuth2, azimuth3)

	// normalisasi ke 0–360
	if azimuth < 0 {
		azimuth += 360
	}

	if azimuth2 < 0 {
		azimuth2 += 360
	}

	if azimuth3 < 0 {
		azimuth3 += 360
	}

	fmt.Printf("azimuth %.2f %.2f %.2f %.2f", azimuth, theta, azimuth2, azimuth3)
	return azimuth
}

func calcKiblat(c *gin.Context) {
	var kibPost KiblatPost

	if err := c.ShouldBindJSON(&kibPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

}
