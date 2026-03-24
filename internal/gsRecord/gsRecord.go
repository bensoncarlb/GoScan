// Struct and handling for incoming document metadata and handling
// TODO rename package
package gsRecord

import (
	"strings"

	"github.com/bensoncb/GoScan/internal/gserrors"
)

// RecordData is a metadata representation of a dcument that has been ingested.
type RecordData struct {
	Size    int               `json:"size"`
	Name    string            `json:"name"`
	Src     string            `json:"src"`
	ImgData []byte            `json:"data"`
	OCRData map[string]string `json:"ocr_data"`
	DocType string            `json:"document_type"`
}

// Setup a new RecordData for a document
func New(Size int, name string, src string, data []byte) (RecordData, error) {
	if strings.TrimSpace(name) == "" {
		return RecordData{}, gserrors.ErrBadParam{Parameter: "Name", Reason: "Missing"}
	} else if strings.TrimSpace(src) == "" {
		return RecordData{}, gserrors.ErrBadParam{Parameter: "Source", Reason: "Missing"}
	} else if len(data) == 0 {
		return RecordData{}, gserrors.ErrBadParam{Parameter: "Data", Reason: "Missing"}
	}

	return RecordData{Name: name, Src: src, ImgData: data}, nil
}
