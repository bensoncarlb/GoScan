package outputFile

import (
	"errors"
	"log"
	"os"
	"path"

	paramerror "github.com/bensoncb/GoScan/internal/errors"
	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

type OutputModule struct {
	Directory string
	IFile     inputFile.InputFile
}

func (o *OutputModule) Save() error {
	log.Printf("recieved output data: %s", o.IFile.Name)

	if len(o.IFile.Name) == 0 {
		return paramerror.ErrBadParam{Parameter: "Name", Reason: "Missing"}
	}
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
		err = paramerror.ErrBadParam{Parameter: "Directory", Reason: "Missing"}
	}

	outModule.Directory = Directory

	if err != nil {
		return outModule, err
	}

	//Check the directory to store received data in exists
	fi, err := os.Stat(Directory)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(Directory, os.ModePerm)
		}
	} else if !fi.IsDir() {
		err = paramerror.ErrBadParam{Parameter: "Directory", Reason: "Directory is a File"}
	}

	return outModule, err
}
