package gsRecord_test

import (
	"errors"
	"testing"

	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/internal/gserrors"
)

func TestBaseGood(t *testing.T) {
	data := []byte("test")

	_, err := gsRecord.New(1, "test", "file", data)

	if err != nil {
		t.Fatalf("Test failed with error: %s", err)
	}
}

func TestBadArgs(t *testing.T) {
	data := []byte("test")
	name := "test"
	iFile, err := gsRecord.New(0, name, "file", data)

	if err == nil {
		t.Fatalf("Missing Size not detected.")
	} else if !errors.Is(err, gserrors.ErrBadParam{}) {
		t.Fatalf("Unexpected error returned: %s", err)
	} else if iFile.Name != name {
		t.Fatalf("Setting Name property failed. Acutal: %s, Expected: %s", iFile.Name, name)
	}
}
