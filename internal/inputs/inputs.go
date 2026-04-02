package inputs

type Module interface {
	Init() error
	Start() error
	Stop() error
	IsReady() (bool, error)
}
