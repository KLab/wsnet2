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
}

func withType(err error, errType ErrType) ErrorWithType {
	if err == nil {
		return nil
	}
	return &errorWithType{err, errType}
}

func (e *errorWithType) ErrType() ErrType {
	return e.errType
}

func (e *errorWithType) Message() string {
	switch e.errType {
	case ErrArgument:
		return "Invalid argument"
	case ErrRoomLimit:
		return "Reached to the max room number"
	case ErrNoJoinableRoom:
		return "No joinable room found"
	case ErrRoomFull:
		return "Room full"
	case ErrAlreadyJoined:
		return "Already exists"
	case ErrNoWatchableRoom:
		return "No watchable room found"
	}
	return ""
}

func (e *errorWithType) Error() string {
	return fmt.Sprintf("%v: %v", e.Message(), e.error.Error())
}

func (e *errorWithType) Unwrap() error {
	return e.error
}

func (e *errorWithType) Format(f fmt.State, c rune) {
	f.Write([]byte(e.Message()))
	f.Write([]byte(": "))
	if m, ok := e.error.(xerrors.Formatter); ok {
		xerrors.FormatError(m, f, c)
	} else {
		f.Write([]byte(e.Error()))
	}
}
