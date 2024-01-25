package allsrv

import (
	"errors"
)

const (
	errTypeUnknown = iota
	errTypeExists
	errTypeInvalid
	errTypeNotFound
	errTypeUnAuthed
	errTypeInternal
)

// Err provides a lightly structured error that we can attach behavior. Additionally,
// the use of fields makes it possible for us to enrich our logging infra without
// blowing up the message cardinality.
type Err struct {
	Type   int
	Msg    string
	Fields []any
}

// Error returns the error message.
func (e Err) Error() string {
	return e.Msg
}

// ExistsErr creates an exists error.
func ExistsErr(msg string, fields ...any) error {
	return Err{
		Type:   errTypeExists,
		Msg:    msg,
		Fields: fields,
	}
}

// NotFoundErr creates a not found error.
func NotFoundErr(msg string, fields ...any) error {
	return Err{
		Type:   errTypeNotFound,
		Msg:    msg,
		Fields: fields,
	}
}

func errFields(err error) []any {
	var aErr Err
	errors.As(err, &aErr)
	return aErr.Fields
}

func IsNotFoundErr(err error) bool {
	return isErrType(err, errTypeNotFound)
}

func IsExistsErr(err error) bool {
	return isErrType(err, errTypeExists)
}

func isErrType(err error, want int) bool {
	var aErr Err
	return errors.As(err, &aErr) && aErr.Type == want
}
