package lobby

import (
	"fmt"

	"golang.org/x/xerrors"
)

type ErrType int

const (
	ErrArgument = ErrType(iota + 1)
	ErrRoomLimit
	ErrNoJoinableRoom
	ErrRoomFull
	ErrAlreadyJoined
	ErrNoWatchableRoom
)

// ErrorWithErrType : ErrTypeとerrorの組
type ErrorWithType interface {
	error
	ErrType() ErrType
	Message() string
}

type errorWithType struct {
	error
	errType ErrType
	msg     string
}

func withType(err error, errType ErrType, msg string) ErrorWithType {
	if err == nil {
		return nil
	}
	return &errorWithType{err, errType, msg}
}

func (e *errorWithType) ErrType() ErrType {
	return e.errType
}

func (e *errorWithType) Message() string {
	return e.msg
}

func (e *errorWithType) Unwrap() error {
	return e.error
}

func (e *errorWithType) Format(f fmt.State, c rune) {
	if m, ok := e.error.(xerrors.Formatter); ok {
		xerrors.FormatError(m, f, c)
	} else {
		f.Write([]byte(e.Error()))
	}
}
