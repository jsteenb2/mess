package allsrv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-metrics"
)

func WithMetrics(mets *metrics.Metrics) func(*serverOpts) {
	return func(o *serverOpts) {
		o.met = mets
	}
}

func WithMux(mux *http.ServeMux) func(*serverOpts) {
	return func(o *serverOpts) {
		o.mux = mux
	}
}

func WithNowFn(fn func() time.Time) func(*serverOpts) {
	return func(o *serverOpts) {
		o.nowFn = fn
	}
}

type ServerV2 struct {
	db DB // 1)

	mux   *http.ServeMux
	mw    func(next http.Handler) http.Handler
	idFn  func() string // 11)
	nowFn func() time.Time
}

func NewServerV2(db DB, opts ...func(*serverOpts)) *ServerV2 {
	opt := serverOpts{
		mux:   http.NewServeMux(),
		idFn:  func() string { return uuid.Must(uuid.NewV4()).String() },
		nowFn: func() time.Time { return time.Now().UTC() },
	}
	for _, o := range opts {
		o(&opt)
	}

	s := ServerV2{
		db:    db,
		mux:   opt.mux,
		idFn:  opt.idFn,
		nowFn: opt.nowFn,
	}

	var mw []func(http.Handler) http.Handler
	if opt.authFn != nil {
		mw = append(mw, opt.authFn)
	}
	mw = append(mw, withTraceID, withStartTime)
	if opt.met != nil { // put metrics last since these are executed LIFO
		mw = append(mw, ObserveHandler("v2", opt.met))
	}
	mw = append(mw, recoverer)

	s.mw = applyMW(mw...)

	s.routes()

	return &s
}

func (s *ServerV2) routes() {
	withContentTypeJSON := applyMW(contentTypeJSON, s.mw)

	// 9)
	s.mux.Handle("POST /v1/foos", withContentTypeJSON(jsonIn(http.StatusCreated, s.createFooV1)))
}

func (s *ServerV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 4)
	s.mux.ServeHTTP(w, r)
}

// API envelope types
type (
	// RespResourceBody represents a JSON-API response body.
	// 	https://jsonapi.org/format/#document-top-level
	//
	// note: data can be either an array or a single resource object. This allows for both.
	RespResourceBody[Attrs any | []any] struct {
		Meta RespMeta         `json:"meta"`
		Errs []RespErr        `json:"errors,omitempty"`
		Data *RespData[Attrs] `json:"data,omitempty"`
	}

	// RespData represents a JSON-API data response.
	//	https://jsonapi.org/format/#document-top-level
	RespData[Attr any | []Attr] struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes Attr   `json:"attributes"`

		// omitting the relationships here for brevity not at lvl 3 RMM
	}

	// RespMeta represents a JSON-API meta object. The data here is
	// useful for our example service. You can add whatever non-standard
	// context that is relevant to your domain here.
	//	https://jsonapi.org/format/#document-meta
	RespMeta struct {
		TookMilli int    `json:"took_ms"`
		TraceID   string `json:"trace_id"`
	}

	// RespErr represents a JSON-API error object. Do note that we
	// aren't implementing the entire error type. Just the most impactful
	// bits for this workshop. Mainly, skipping Title & description separation.
	//	https://jsonapi.org/format/#error-objects
	RespErr struct {
		Status int            `json:"status,string"`
		Code   int            `json:"code"`
		Msg    string         `json:"message"`
		Source *RespErrSource `json:"source"`
	}

	// RespErrSource represents a JSON-API err source.
	//	https://jsonapi.org/format/#error-objects
	RespErrSource struct {
		Pointer   string `json:"pointer"`
		Parameter string `json:"parameter,omitempty"`
		Header    string `json:"header,omitempty"`
	}
)

type (
	// ReqCreateFooV1 represents the request body for the create foo API.
	ReqCreateFooV1 struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}

	// FooAttrs are the attributes for foo data.
	FooAttrs struct {
		Name      string `json:"name"`
		Note      string `json:"note"`
		CreatedAt string `json:"created_at"`
	}
)

func (s *ServerV2) createFooV1(ctx context.Context, req ReqCreateFooV1) (RespData[FooAttrs], []RespErr) {
	newFoo := Foo{
		ID:        s.idFn(),
		Name:      req.Name,
		Note:      req.Note,
		CreatedAt: s.nowFn(),
	}
	if err := s.db.CreateFoo(ctx, newFoo); err != nil {
		return RespData[FooAttrs]{}, toRespErrs(err)
	}

	out := newFooData(newFoo.ID, FooAttrs{
		Name:      newFoo.Name,
		Note:      newFoo.Note,
		CreatedAt: toTimestamp(newFoo.CreatedAt),
	})
	return out, nil
}

func newFooData(id string, attrs FooAttrs) RespData[FooAttrs] {
	return RespData[FooAttrs]{
		Type:       "foo",
		ID:         id,
		Attributes: attrs,
	}
}

func toTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

func jsonIn[ReqBody, Attr any](successCode int, fn func(context.Context, ReqBody) (RespData[Attr], []RespErr)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			reqBody ReqBody
			errs    []RespErr
			out     *RespData[Attr]
		)
		if respErr := decodeReq(r, &reqBody); respErr != nil {
			errs = append(errs, *respErr)
		} else {
			var data RespData[Attr]
			data, errs = fn(r.Context(), reqBody)
			if len(errs) == 0 {
				out = &data
			}
		}

		status := successCode
		for _, e := range errs {
			if e.Status > status {
				status = e.Status
			}
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(RespResourceBody[Attr]{
			Meta: getMeta(r.Context()),
			Errs: errs,
			Data: out,
		}) // 10.b)
	})
}

func decodeReq(r *http.Request, v any) *RespErr {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respErr := RespErr{
			Status: http.StatusBadRequest,
			Msg:    "failed to decode request body: " + err.Error(),
			Source: &RespErrSource{
				Pointer: "/data",
			},
			Code: errTypeInvalid,
		}
		if unmarshErr := new(json.UnmarshalTypeError); errors.As(err, &unmarshErr) {
			respErr.Source.Pointer += "/attributes/" + unmarshErr.Field
		}
		return &respErr
	}

	return nil
}

func toRespErrs(err error) []RespErr {
	if e := new(Err); errors.As(err, e) {
		return []RespErr{{
			Code: errCode(e),
			Msg:  e.Msg,
		}}
	}

	errs, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return nil
	}

	var out []RespErr
	for _, e := range errs.Unwrap() {
		out = append(out, toRespErrs(e)...)
	}

	return out
}

func errCode(err *Err) int {
	switch err.Type {
	case errTypeExists:
		return http.StatusConflict
	case errTypeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// WithBasicAuthV2 sets the authorization fn for the server to basic auth.
// 3)
func WithBasicAuthV2(adminUser, adminPass string) func(*serverOpts) {
	return func(s *serverOpts) {
		s.authFn = func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if user, pass, ok := r.BasicAuth(); !(ok && user == adminUser && pass == adminPass) {
					w.WriteHeader(http.StatusUnauthorized) // 9)
					json.NewEncoder(w).Encode(RespResourceBody[any]{
						Meta: getMeta(r.Context()),
						Errs: []RespErr{{
							Status: http.StatusUnauthorized,
							Code:   errTypeUnAuthed,
							Msg:    "unauthorized access",
							Source: &RespErrSource{
								Header: "Authorization",
							},
						}},
					})
					return
				}
				next.ServeHTTP(w, r)
			})
		}
	}
}

func contentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(RespResourceBody[any]{
				Meta: getMeta(r.Context()),
				Errs: []RespErr{{
					Code: http.StatusUnsupportedMediaType,
					Msg:  "received invalid media type",
				}},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getMeta(ctx context.Context) RespMeta {
	return RespMeta{
		TookMilli: int(took(ctx).Milliseconds()),
		TraceID:   getTraceID(ctx),
	}
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr == nil {
				return
			}

			if rvr == http.ErrAbortHandler {
				// we don't recover http.ErrAbortHandler so the response
				// to the client is aborted, this should not be logged
				panic(rvr)
			}

			w.WriteHeader(http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

const (
	ctxStartTime = "start"
	ctxTraceID   = "trace-id"
)

func withTraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Mess-Trace-Id")
		if traceID == "" {
			traceID = uuid.Must(uuid.NewV4()).String()
		}
		ctx := context.WithValue(r.Context(), ctxTraceID, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value(ctxTraceID).(string)
	return traceID
}

func withStartTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxStartTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func took(ctx context.Context) time.Duration {
	t, _ := ctx.Value(ctxStartTime).(time.Time)
	return time.Since(t)
}

func applyMW[T any](fns ...func(T) T) func(T) T {
	return func(v T) T {
		for i := len(fns) - 1; i >= 0; i-- {
			v = fns[i](v)
		}
		return v
	}
}