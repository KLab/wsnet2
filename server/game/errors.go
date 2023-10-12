package game

import (
	"errors"
	"fmt"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
)

// ErrorWithCode : gRPCのコードとerrorの組
type ErrorWithCode interface {
	error
	Code() codes.Code
	IsNormal() bool
}

type errorWithCode struct {
	error
	code     codes.Code
	isNormal bool
}

func WithCode(err error, code codes.Code) ErrorWithCode {
	if err == nil {
		return nil
	}
	var ewc ErrorWithCode
	if errors.As(err, &ewc) {
		return errorWithCode{err, code, ewc.IsNormal()}
	}

	return errorWithCode{err, code, false}
}

func NormalWithCode(err error, code codes.Code) ErrorWithCode {
	if err == nil {
		return nil
	}
	return errorWithCode{err, code, true}
}

func (e errorWithCode) Code() codes.Code {
	return e.code
}

func (e errorWithCode) IsNormal() bool {
	return e.isNormal
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
