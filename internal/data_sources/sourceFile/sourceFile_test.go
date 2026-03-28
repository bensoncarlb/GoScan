package sourceFile_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/bensoncarlb/GoScan/internal/data_sources/sourceFile"
	"github.com/bensoncarlb/GoScan/internal/outputs/outputFile"
	"github.com/bensoncarlb/GoScan/internal/server"
)

/*
* Test basic initialization and file pick up
 */
func TestInit(t *testing.T) {
	OutDir := t.TempDir()
	InDir := t.TempDir()

	t.Logf("In: %s, Out: %s", InDir, OutDir)

	svr := StartServer(OutDir)
	defer svr.Stop()

	src, err := sourceFile.New(InDir, "http://localhost:8090/data")
	if err != nil {
		t.Fatalf("SourceFile init failed: %s", err)
	}

	err = src.Start()

	if err != nil {
		t.Fatalf("Start failed: %s", err)
	}

	defer src.Stop()

	fil, err := os.Create(path.Join(InDir, "TestInit"))

	if err != nil {
		t.Fatalf("Failed to create test dir: %s", err)
	}

	fil.Write([]byte("TestInit"))
	fil.Close()

	time.Sleep(time.Second)

	_, err = os.Stat(path.Join(OutDir, "TestInit"))

	if err != nil {
		t.Fatalf("File check failed: %s", err)
	}
}

/*
* Spin up a new processing server (server.go)
* to watch the provided TempDir for testing files
 */
func StartServer(OutDir string) *server.Server {
	//Setup handler for outputing final data
	ModOutput, err := outputFile.New(OutDir)

	if err != nil {
		panic(err)
	}

	//Setup listening server
	svr := server.Server{ModOutput: &ModOutput}
	/*
		if err := svr.New(); err != nil {
			panic(err)
		}

		if err := svr.Start(); err != nil {
			panic(err)
		}
	*/
	//TODO
	return &svr
}
