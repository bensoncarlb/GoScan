package documentType

import (
	"image"
)

type DocumentType struct {
	Title      string           `json:"title"`
	Identifier string           `json:"identifier"`
	Regions    []DocumentRegion `json:"regions"`
}

type DocumentRegion struct {
	RegionTitle string          `json:"region_title"`
	FieldName   string          `json:"data_field"`
	Region      image.Rectangle `json:"region"`
}
