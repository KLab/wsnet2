package lobby

import (
	"fmt"

	"golang.org/x/xerrors"
)

// ErrorWithStatus : HTTPステータスコードとerrorの組
type ErrorWithStatus interface {
	error
	Status() int
	Message() string
}

type errorWithStatus struct {
	error
	status int
	msg    string
}

func withStatus(err error, status int, msg string) ErrorWithStatus {
	if err == nil {
		return nil
	}
	return &errorWithStatus{err, status, msg}
}

func (e *errorWithStatus) Status() int {
	return e.status
}

func (e *errorWithStatus) Message() string {
	return e.msg
}

func (e *errorWithStatus) Unwrap() error {
	return e.error
}

func (e *errorWithStatus) Format(f fmt.State, c rune) {
	if m, ok := e.error.(xerrors.Formatter); ok {
		xerrors.FormatError(m, f, c)
	} else {
		f.Write([]byte(e.Error()))
	}
}
