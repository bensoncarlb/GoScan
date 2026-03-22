// Handle reading data from a provided document using OCR
package ocr

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"log"
	"os/exec"

	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

// For a provided item, read and return the OCR'd data
func ReadRegion(i *[]byte) (string, error) {
	data := base64.StdEncoding.EncodeToString(*i)

	cmd := fmt.Sprintf("echo %s | base64 -d | tesseract stdin stdout", data)

	res, err := exec.Command("bash", "-c", cmd).Output()

	return fmt.Sprintf("%s", res), err
}

/*
* Attempt to identify the provided image
 */
func FormIdentify(d *inputFile.InputFile) error {
	//TODO implement
	if d.DocType != "" {
		return fmt.Errorf("Document already identified as %v", d.DocType)
	}

	if len(d.ImgData) == 0 {
		return fmt.Errorf("No data provided")
	}

	d.DocType = "Test"

	return nil
}

func ConvertToGray(img image.Image) *image.Gray {
	// 1. Create a new blank image.Gray with the same bounds as the original image.
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	log.Printf("img convert")
	// 2. Draw the original image onto the new grayscale image.
	// The draw.Src operation uses the destination's color model to convert the source pixels.
	draw.Draw(gray, bounds, img, bounds.Min, draw.Src)

	return gray
}

// func ConvertToGray(img image.Image) *image.Gray {
// 	var (
// 		bounds = img.Bounds()
// 		gray   = image.NewGray(bounds)
// 	)

// 	for x := 0; x < bounds.Max.X; x++ {
// 		for y := 0; y < bounds.Max.Y; y++ {
// 			var rgba = img.At(x, y)
// 			gray.Set(x, y, rgba)
// 		}
// 	}
// 	return gray
// }
