package game

import (
	"fmt"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
)

// ErrorWithCode : gRPCのコードとerrorの組
type ErrorWithCode interface {
	error
	Code() codes.Code
}

type errorWithCode struct {
	error
	code codes.Code
}

func WithCode(err error, code codes.Code) ErrorWithCode {
	if err == nil {
		return nil
	}
	return errorWithCode{err, code}
}

func (e errorWithCode) Code() codes.Code {
	return e.code
}

func (e errorWithCode) Unwrap() error {
	return e.error
}

func (e errorWithCode) Format(f fmt.State, c rune) {
	if m, ok := e.error.(xerrors.Formatter); ok {
		xerrors.FormatError(m, f, c)
	} else {
		f.Write([]byte(e.Error()))
	}
}
