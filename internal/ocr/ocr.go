package ocr

import (
	"fmt"

	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

func ReadImage(i *[]byte) (string, error) {

	return "test", nil
}

func FormIdentify(d *inputFile.InputFile) error {
	if d.DocType != "" {
		return fmt.Errorf("Document already identified as %v", d.DocType)
	}

	if len(d.Data) == 0 {
		return fmt.Errorf("No data provided")
	}

	d.DocType = "Test"

	return nil
}
