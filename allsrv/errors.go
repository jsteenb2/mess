package allsrv

import (
	"github.com/jsteenb2/errors"
)

const (
	ErrKindExists   = errors.Kind("exists")
	ErrKindInvalid  = errors.Kind("invalid")
	ErrKindNotFound = errors.Kind("not found")
	ErrKindUnAuthed = errors.Kind("unauthorized")
	ErrKindInternal = errors.Kind("internal")
)

const (
	errCodeExist    = 1
	errCodeInvalid  = 2
	errCodeNotFound = 3
	errCodeUnAuthed = 4
	errCodeInternal = 5
)

func errCode(kind error) int {
	switch {
	case errors.Is(kind, ErrKindExists):
		return errCodeExist
	case errors.Is(kind, ErrKindInvalid):
		return errCodeInvalid
	case errors.Is(kind, ErrKindNotFound):
		return errCodeNotFound
	case errors.Is(kind, ErrKindUnAuthed):
		return errCodeUnAuthed
	case errors.Is(kind, ErrKindInternal):
		return errCodeInternal
	default:
		return errCode(ErrKindInternal)
	}
}

var (
	errIDRequired = InvalidErr("id is required")
)

// ExistsErr creates an exists error.
func ExistsErr(msg string, fields ...any) error {
	return errors.New(msg, errors.KVs(fields...), ErrKindExists, errors.SkipCaller)
}

func InvalidErr(msg string, fields ...any) error {
	return errors.New(msg, errors.KVs(fields...), ErrKindInvalid, errors.SkipCaller)
}

func InternalErr(msg string, fields ...any) error {
	return errors.New(msg, errors.KVs(fields...), ErrKindInternal, errors.SkipCaller)
}

// NotFoundErr creates a not found error.
func NotFoundErr(msg string, fields ...any) error {
	return errors.New(msg, errors.KVs(fields...), ErrKindNotFound, errors.SkipCaller)
}

func unauthedErr(msg string, fields ...any) error {
	return errors.New(msg, errors.KVs(fields...), ErrKindUnAuthed, errors.SkipCaller)
}
