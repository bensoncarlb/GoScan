package outputFile

import (
	"fmt"
	"os"
	"path"

	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

type OutputModule struct {
	Directory string
	IFile     inputFile.InputFile
}

type ErrBadParam struct {
	Parameter string
	Reason    string
}

func (e ErrBadParam) Error() string {
	return fmt.Sprintf("Invalid parameter %s, Reason: %s", e.Parameter, e.Reason)
}

func (e ErrBadParam) Is(err error) bool {
	_, ok := err.(ErrBadParam)
	return ok
}

func (o *OutputModule) Save() error {
	println("recieved output data")
	//Save off received data
	fil, err := os.Create(path.Join(o.Directory, o.IFile.Name))

	if err != nil {
		return err
	}

	defer fil.Close()

	fil.Write(o.IFile.Data)

	return nil
}

func New(Directory string) (OutputModule, error) {
	outModule := OutputModule{}
	var err error = nil

	if len(Directory) == 0 {
		err = ErrBadParam{Parameter: "Directory", Reason: "Missing"}
	}

	outModule.Directory = Directory

	return outModule, err
}
