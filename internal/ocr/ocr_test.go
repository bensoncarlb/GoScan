package ocr_test

import (
	"testing"

	"github.com/bensoncb/GoScan/internal/gsRecord"
	"github.com/bensoncb/GoScan/internal/ocr"
)

/*
* Check a basic valid case
 */
func TestValid(t *testing.T) {
	iFile, err := gsRecord.New(1, "test", "file", []byte("test"))

	if err != nil {
		t.Fatalf("Failed to setup inputFile: %s", err)
	}

	err = ocr.FormIdentify(&iFile)

	if err != nil {
		t.Fatalf("Form Identify failed: %s", err)
	}

	res, err := ocr.ReadRegion(iFile.ImgData)

	if err != nil {
		t.Fatalf("Error during data read: %s", err)
	} else if res != "test" {
		t.Fatalf("Form data read failed. Expected: %s, Got: %s", "test", res)
	}

}
