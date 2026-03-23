// Output module for saving result to file storage
package outputFile

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path"
	"strings"

	"github.com/bensoncb/GoScan/internal/gsRecord"
	"github.com/bensoncb/GoScan/internal/gserrors"
)

type OutputModule struct {
	Directory string
	IFile     gsRecord.RecordData
}

// Take the provided OutputModule.IFile and save the data to a file location
func (o *OutputModule) Save() error {
	log.Printf("Recieved output data: %s", o.IFile.Name)

	if strings.TrimSpace(o.IFile.Name) == "" {
		//TODO enumerate reasons
		return gserrors.ErrBadParam{Parameter: "Name", Reason: "Missing"}
	}

	//Save off received data
	filImg, err := os.Create(path.Join(o.Directory, o.IFile.Name))

	if err != nil {
		return err
	}

	defer filImg.Close()

	n, err := filImg.Write(o.IFile.ImgData)

	if err != nil {
		return err
	}

	log.Printf("Wrote image file containing %d bytes", n)

	filData, err := os.Create(path.Join(o.Directory, o.IFile.Name+".txt"))

	if err != nil {
		return err
	}

	defer filData.Close()

	data, err := json.Marshal(o.IFile)

	if err != nil {
		panic(err)
	}

	n, err = filData.Write(data)

	if err != nil {
		panic(err)
	}

	log.Printf("Wrote ocr data file containing %d bytes", n)

	return nil
}

// Setup a new output module targeted to the provided Directory
func New(Directory string) (OutputModule, error) {
	if strings.TrimSpace(Directory) == "" {
		return OutputModule{}, gserrors.ErrBadParam{Parameter: "Directory", Reason: "Missing"}
	}

	outModule := OutputModule{Directory: Directory}

	//Check the directory to store received data in exists
	fi, err := os.Stat(Directory)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(Directory, os.ModePerm)
		}

		if err != nil {
			return OutputModule{}, err
		}
	} else if !fi.IsDir() {
		return OutputModule{}, gserrors.ErrBadParam{Parameter: "Directory", Reason: "Directory is a File"}
	}

	return outModule, err
}
