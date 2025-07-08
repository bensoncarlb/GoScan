package structs

type InputFile struct {
	Size int    `json:"size"`
	Name string `json:"name"`
	Src  string `json:"src"`
	Data []byte `json:"data"`
}
