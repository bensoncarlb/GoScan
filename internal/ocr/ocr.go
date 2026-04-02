// Handle reading data from a provided document using OCR
package ocr

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os/exec"
)

func init() {
	cmd := exec.Command("tesseract", "--version")

	_, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal(err)
	}
}

// For a provided item, read and return the OCR'd data
func ReadRegion(img *image.Gray, r image.Rectangle) (string, error) {
	subImg := new(bytes.Buffer)
	err := png.Encode(subImg, img.SubImage(r))

	if err != nil {
		return "", err
	}

	cmd := exec.Command("tesseract", "stdin", "stdout")

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return "", err
	}

	stdin.Write(subImg.Bytes())
	stdin.Close()

	res, err := cmd.CombinedOutput()

	if err != nil {
		return "", err
	}

	return string(res), nil
}

// Convert an image.Image to image.Gray to access the SubImage method
func ConvertToGray(imgData []byte) *image.Gray {
	//TODO Check DecodeConfig
	img, _, err := image.Decode(bytes.NewReader(imgData))

	if err != nil {
		log.Fatalf("Failed reading image: %s", err)
	}

	gray := image.NewGray(img.Bounds())

	draw.Draw(gray, img.Bounds(), img, img.Bounds().Min, draw.Src)

	return gray
}
