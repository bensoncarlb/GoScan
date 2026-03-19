// Handle reading data from a provided document using OCR
package ocr

import (
	"fmt"

	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

/*
* For a provided item, read and return the OCR'd data
 */
func ReadImage(i *[]byte) (string, error) {
	//TODO implement
	return "test", nil
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
