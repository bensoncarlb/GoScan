package outputs

import (
	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/structs"
)

type Module interface {
	Init() error
	Save(*gsRecord.RecordData) error
	ListItems() (structs.RspGetItems, error)
	Retrieve(string) (*gsRecord.RecordData, error)
}
