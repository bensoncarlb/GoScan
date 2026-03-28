// Handle reading data from a provided document using OCR
package ocr

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os/exec"
)

func init() {
	pipeErr, wpipeErr := io.Pipe()
	defer pipeErr.Close()
	//TODO make work
	cmd := exec.Command("tesseracst", "--version")
	cmd.Stdout = wpipeErr
	cmd.Run()
	wpipeErr.Close()

	if res, err := io.ReadAll(pipeErr); err != nil {
		panic(err)
	} else if len(res) > 0 {
		panic(res)
	}
}

// For a provided item, read and return the OCR'd data
func ReadRegion(img *image.Gray, r image.Rectangle) (string, error) {
	subImg := new(bytes.Buffer)
	err := png.Encode(subImg, img.SubImage(r))

	if err != nil {
		return "", err
	}

	pipeRes, wpipeRes := io.Pipe()
	pipeErr, wpipeErr := io.Pipe()

	defer pipeRes.Close()
	defer pipeErr.Close()

	cmd := exec.Command("tesseract", "stdin", "stdout")
	cmd.Stdin = bytes.NewReader(subImg.Bytes())

	cmd.Stdout = wpipeRes
	cmd.Stderr = wpipeErr

	if err = cmd.Run(); err != nil {
		return "", err
	}

	wpipeRes.Close()
	wpipeErr.Close()

	if rd, err := io.ReadAll(pipeErr); len(rd) > 0 {
		return "", err
	}

	res, err := io.ReadAll(pipeRes)

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
