// Output module for saving result to file storage
package outputFile

import (
	"encoding/json"
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

/*
* Take the provided OutputModule.IFile and save the data to a file location
 */
func (o *OutputModule) Save() error {
	log.Printf("Recieved output data: %s", o.IFile.Name)

	if len(o.IFile.Name) == 0 {
		return paramerror.ErrBadParam{Parameter: "Name", Reason: "Missing"}
	}

	//Save off received data
	fil, err := os.Create(path.Join(o.Directory, o.IFile.Name))

	if err != nil {
		return err
	}

	defer fil.Close()

	fil.Write(o.IFile.ImgData)

	fil, err = os.Create(path.Join(o.Directory, o.IFile.Name+".txt"))

	if err != nil {
		return err
	}

	defer fil.Close()

	data, err := json.Marshal(o.IFile)

	if err != nil {
		panic(err)
	}

	_, err = fil.Write(data)

	if err != nil {
		panic(err)
	}

	return nil
}

/*
* Setup a new output module targeted to the provided Directory
 */
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
