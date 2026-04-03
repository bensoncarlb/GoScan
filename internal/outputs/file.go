// Output module for saving result to file storage
package outputs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/internal/gserrors"
	"github.com/bensoncarlb/GoScan/structs"
)

type OutputFile struct {
	Directory string
}

// Take the provided OutputModule.IFile and save the data to a file location
func (o OutputFile) Save(record *gsRecord.RecordData) error {
	log.Printf("Recieved output data: %s", record.Name)

	if strings.TrimSpace(record.Name) == "" {
		//TODO enumerate reasons
		return gserrors.ErrBadParam{Parameter: "Name", Reason: "Missing"}
	}

	//Save off received data
	filImg, err := os.Create(path.Join(o.Directory, record.Name))

	if err != nil {
		return err
	}

	defer filImg.Close()

	n, err := filImg.Write(record.ImgData)

	if err != nil {
		return err
	}

	log.Printf("Wrote image file containing %d bytes", n)

	filData, err := os.Create(path.Join(o.Directory, record.Name+".txt"))

	if err != nil {
		return err
	}

	defer filData.Close()

	data, err := json.Marshal(record)

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

func (o OutputFile) validate() error {
	if strings.TrimSpace(o.Directory) == "" {
		return gserrors.ErrBadParam{Parameter: "Directory", Reason: "Missing"}
	}

	return nil
}

func (o OutputFile) Init() error {
	err := o.validate()

	if err != nil {
		return err
	}

	fi, err := os.Stat(o.Directory)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(o.Directory, os.ModePerm)
		}

		if err != nil {
			return err
		}
	} else if !fi.IsDir() {
		return gserrors.ErrBadParam{Parameter: "Directory", Reason: "Directory is a File"}
	}

	return nil
}

func (o OutputFile) ListItems() (structs.RspGetItems, error) {
	if strings.TrimSpace(o.Directory) == "" {
		return structs.RspGetItems{}, errors.New("no output Directory configured")
	}

	dir, err := os.ReadDir(o.Directory)

	if err != nil {
		return structs.RspGetItems{}, fmt.Errorf("failed to open output directory %s", o.Directory)
	}

	dirFiles := make([]string, 0, len(dir))

	for _, dirEntry := range dir {
		if !dirEntry.IsDir() {
			dirFiles = append(dirFiles, dirEntry.Name())
		}
	}

	return structs.RspGetItems{Items: slices.Clip(dirFiles)}, nil
}

func (o OutputFile) Retrieve(itemName string) (*gsRecord.RecordData, error) {
	fil, err := os.OpenInRoot(o.Directory, itemName)

	if err != nil {
		return &gsRecord.RecordData{}, err
	}

	record := gsRecord.RecordData{}
	json.NewDecoder(fil).Decode(&record)

	return &record, nil
}
