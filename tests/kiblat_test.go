package tests

import (
	"alkaukaba-backend/controllers"
	"fmt"
	"testing"
)

func Test_calcKiblat(t *testing.T) {
	cek := controllers.ArahKiblat(-7.3390, 112.6208)
	fmt.Printf("\ncek atan %.2f", cek)
}
