package paramerror

import "fmt"

type ErrBadParam struct {
	Parameter string
	Reason    string
}

func (e ErrBadParam) Error() string {
	return fmt.Sprintf("Invalid Parameter: %s, Reason: %s", e.Parameter, e.Reason)
}

func (e ErrBadParam) Is(err error) bool {
	_, ok := err.(ErrBadParam)
	return ok
}
