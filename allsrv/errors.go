package allsrv

const (
	errTypeExists   = "exists"
	errTypeNotFound = "not found"
)

// Err provides a lightly structured error that we can attach behavior. Additionally,
// the use of fields makes it possible for us to enrich our logging infra without
// blowing up the message cardinality.
type Err struct {
	Type   string
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
