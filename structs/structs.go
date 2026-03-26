package structs

import (
	"image"
	"strings"
)

// Struct for the Request to the server to delete a Document Type
type ReqDeleteDocumentType struct {
	DocumentType string `json:"document_type"`
}

// Struct for the Response to a request for listing configured Document Types
type RspGetDocumentTypes struct {
	DocumentTypes []DocumentType `json:"document_types"`
}

// Struct for the Response to a request for listing currently processed Items
type RspGetItems struct {
	Items []string `json:"items"`
}

// Struct for the Request
type ReqRetrieveItem struct {
	ItemName string `json:"item_name"`
}

type RspRetrieveItem struct {
	Fields  map[string]string `json:"fields"`
	ImgData []byte            `json:"img_data"`
}

type DocumentType struct {
	Title      string           `json:"title"`
	Identifier string           `json:"identifier"`
	Regions    []DocumentRegion `json:"regions"`
}

func (d *DocumentType) IsValid() bool {
	if strings.TrimSpace(d.Title) == "" {
		return false
	} else if strings.TrimSpace(d.Identifier) == "" {
		return false
	} else if len(d.Regions) == 0 {
		return false
	} else {
		for _, reg := range d.Regions {
			if !reg.IsValid() {
				return false
			}
		}
	}

	return true
}

type DocumentRegion struct {
	RegionTitle string          `json:"region_title"`
	FieldName   string          `json:"data_field"`
	Region      image.Rectangle `json:"region"`
}

func (d *DocumentRegion) IsValid() bool {
	if strings.TrimSpace(d.FieldName) == "" {
		return false
	} else if strings.TrimSpace(d.RegionTitle) == "" {
		return false
	} else if image.Rectangle.Empty(d.Region) {
		return false
	}

	return true
}
