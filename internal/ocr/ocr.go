// Handle reading data from a provided document using OCR
package ocr

import (
	"encoding/base64"
	"fmt"
	"os/exec"

	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

/*
* For a provided item, read and return the OCR'd data
 */
func ReadImage(i *[]byte) (string, error) {
	//TODO implement
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

	if len(d.Data) == 0 {
		return fmt.Errorf("No data provided")
	}

	d.DocType = "Test"

	return nil
}
