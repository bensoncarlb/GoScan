// Struct and handling for incoming document metadata and handling
package inputFile

import (
	"image"

	paramerror "github.com/bensoncb/GoScan/internal/errors"
)

// InputFile is a metadata representation of a dcument that has been ingested.
type InputFile struct {
	Size    int               `json:"size"`
	Name    string            `json:"name"`
	Src     string            `json:"src"`
	ImgData []byte            `json:"data"`
	OCRData map[string]string `json:"ocr_data"`
	SizeX   int32
	SizeY   int32
	DocType string // TODO enumerate this
	State   string // TODO enumerate this
	Regions map[string]image.Rectangle
}

/*
* Setup a new InputFile for a document
 */
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
	iFile.ImgData = Data

	return iFile, err
}
