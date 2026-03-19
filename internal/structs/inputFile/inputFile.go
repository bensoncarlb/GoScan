// Package inputFile does X
package inputFile

import (
	paramerror "github.com/bensoncb/GoScan/internal/errors"
)

// InputFile is a metadata representation of a file that has been ingested.
type InputFile struct {
	Size    int    `json:"size"`
	Name    string `json:"name"`
	Src     string `json:"src"`
	Data    []byte `json:"data"`
	SizeX   int32
	SizeY   int32
	DocType string // TODO enumerate this
	State   string // TODO enumerate this
	//Regions []Region
}

type Region struct {
	StartX int32
	StartY int32
	EndX   int32
	EndY   int32
}

func New(Size int, Name string, Data []byte) (InputFile, error) {
	iFile := InputFile{}
	var err error = nil

	if Size <= 0 {
		err = paramerror.ErrBadParam{Parameter: "Size"}
	} else if len(Name) == 0 {
		return iFile, paramerror.ErrBadParam{Parameter: "Name"}
	} else if len(Data) == 0 {
		return iFile, paramerror.ErrBadParam{Parameter: "Data"}
	}

	iFile.Size = Size
	iFile.Name = Name
	iFile.Data = Data

	return iFile, err
}
