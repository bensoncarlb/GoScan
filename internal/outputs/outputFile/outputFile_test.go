package outputFile_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bensoncb/GoScan/internal/outputs/outputFile"
	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

/*
* Test a basic valid scenario
 */
func TestGoodOutput(t *testing.T) {
	TestFile := "TestInit"
	TestData := []byte("TestInit")
	TestDir := t.TempDir()
	OutputModule, err := outputFile.New(TestDir)

	if err != nil {
		t.Fatalf("Failed module setup: %s", err)
	}

	OutputModule.IFile, err = inputFile.New(1, TestFile, TestData)

	if err != nil {
		t.Fatalf("Failed to setup test file: %s", err)
	}

	err = OutputModule.Save()

	if err != nil {
		t.Fatalf(("Failed to save data: %s"), err)
	}

	file, err := os.ReadFile(filepath.Join(TestDir, TestFile))

	if err != nil {
		t.Fatalf("Failed opening file: %s", err)
	}

	if !bytes.Equal(file, TestData) {
		t.Fatalf("Data mismatch. Expected: %s, Got: %s", TestData, file)
	}
}
