package outputs_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/internal/outputs"
)

/*
* Test a basic valid scenario
 */
func TestGoodOutput(t *testing.T) {
	TestFile := "TestInit"
	TestData := []byte("TestInit")
	TestDir := t.TempDir()

	outputModule := outputs.OutputFile{Directory: TestDir}
	err := outputModule.Init()

	if err != nil {
		t.Fatalf("Failed module setup: %s", err)
	}

	record, err := gsRecord.New(1, TestFile, "file", TestData)

	if err != nil {
		t.Fatalf("Failed to setup test file: %s", err)
	}
	err = outputModule.Save(&record)

	if err != nil {
		t.Fatalf(("Failed to save data: %s"), err)
	}

	path, err := os.OpenInRoot(TestDir, TestFile)

	if err != nil {
		t.Fatalf("Failed opening file: %s", err)
	}

	file, err := io.ReadAll(path)

	if err != nil {
		t.Fatalf("Failed opening file: %s", err)
	}

	if !bytes.Equal(file, TestData) {
		t.Fatalf("Data mismatch. Expected: %s, Got: %s", TestData, file)
	}
}
