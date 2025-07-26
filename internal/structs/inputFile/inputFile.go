// Package inputFile does X
package inputFile

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
