package documentType

import (
	"image"
	"strings"
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

func (d *DocumentRegion) IsValid() bool {
	if strings.TrimSpace(d.FieldName) == "" {
		return false
	} else if strings.TrimSpace(d.RegionTitle) == "" {
		return false
	} /*else if d.Region == image.Rectangle{} {
		return false
	} */

	return true
}
