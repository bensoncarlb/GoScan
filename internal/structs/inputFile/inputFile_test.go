package inputFile

import (
	"errors"
	"testing"
)

func TestBaseGood(t *testing.T) {
	data := []byte("test")

	_, err := New(1, "test", data)

	if err != nil {
		t.Fatalf("Test failed with error: %s", err)
	}
}

func TestBadArgs(t *testing.T) {
	data := []byte("test")
	name := "test"
	iFile, err := New(0, name, data)

	if err == nil {
		t.Fatalf("Missing Size not detected.")
	} else if !errors.Is(err, ErrBadParam{}) {
		t.Fatalf("Unexpected error returned: %s", err)
	} else if iFile.Name != name {
		t.Fatalf("Setting Name property failed. Acutal: %s, Expected: %s", iFile.Name, name)
	}
}
