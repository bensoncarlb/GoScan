// Handle reading data from a provided document using OCR
package ocr

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os/exec"
)

func init() {
	//TODO check tesseract
}

// For a provided item, read and return the OCR'd data
func ReadRegion(img *image.Gray, r image.Rectangle) (string, error) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img.SubImage(r))

	if err != nil {
		return "", err
	}

	data := base64.StdEncoding.EncodeToString(buf.Bytes())
	//TODO handle better
	/*cmd := exec.command
	cmd.stdin ...
	cmd.run
	cmd.stdout
	cmd.stderr*/
	cmd := fmt.Sprintf("echo %s | base64 -d | tesseract stdin stdout", data)

	res, err := exec.Command("bash", "-c", cmd).Output()

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
