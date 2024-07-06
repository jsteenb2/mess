package allsrv

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-metrics"
	"github.com/jsteenb2/errors"
)

type SvrOptFn func(o *serverOpts)

func WithMetrics(mets *metrics.Metrics) SvrOptFn {
	return func(o *serverOpts) {
		o.met = mets
	}
}

func WithMux(mux *http.ServeMux) SvrOptFn {
	return func(o *serverOpts) {
		o.mux = mux
	}
}

type ServerV2 struct {
	mux *http.ServeMux
	svc SVC
	mw  func(next http.Handler) http.Handler
}

func NewServerV2(svc SVC, opts ...SvrOptFn) *ServerV2 {
	opt := serverOpts{
		mux: http.NewServeMux(),
	}
	for _, o := range opts {
		o(&opt)
	}

	s := ServerV2{
		svc: svc,
		mux: opt.mux,
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
	s.mux.Handle("POST /v1/foos", withContentTypeJSON(jsonIn(resourceTypeFoo, http.StatusCreated, s.createFooV1)))
	s.mux.Handle("GET /v1/foos/{id}", s.mw(read(s.readFooV1)))
	s.mux.Handle("PATCH /v1/foos/{id}", withContentTypeJSON(jsonIn(resourceTypeFoo, http.StatusOK, s.updateFooV1)))
	s.mux.Handle("DELETE /v1/foos/{id}", s.mw(del(s.delFooV1)))
}

func (s *ServerV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 4)
	s.mux.ServeHTTP(w, r)
}

// API envelope types
type (
	// RespBody represents a JSON-API response body.
	// 	https://jsonapi.org/format/#document-top-level
	//
	// note: data can be either an array or a single resource object. This allows for both.
	RespBody[Attr Attrs] struct {
		Meta RespMeta    `json:"meta"`
		Errs []RespErr   `json:"errors,omitempty"`
		Data *Data[Attr] `json:"data,omitempty"`
	}

	// Attrs can be either a document or a collection of documents.
	Attrs interface {
		any | []Attrs
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

	// ReqBody represents a JSON-API request body.
	//	https://jsonapi.org/format/#crud-creating
	ReqBody[Attr Attrs] struct {
		Data Data[Attr] `json:"data"`
	}
)

// Data represents a JSON-API data response.
//
//	https://jsonapi.org/format/#document-top-level
type Data[Attr Attrs] struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Attrs Attr   `json:"attributes"`

	// omitting the relationships here for brevity not at lvl 3 RMM
}

func (d Data[Attr]) getType() string {
	return d.Type
}

const (
	resourceTypeFoo = "foo"
)

type (
	ReqCreateFooV1 = ReqBody[FooCreateAttrs]

	FooCreateAttrs struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}

	// ResourceFooAttrs are the attributes of a foo resource.
	ResourceFooAttrs struct {
		Name      string `json:"name"`
		Note      string `json:"note"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
)

func (s *ServerV2) createFooV1(ctx context.Context, req ReqCreateFooV1) (*Data[ResourceFooAttrs], []RespErr) {
	newFoo, err := s.svc.CreateFoo(ctx, Foo{
		Name: req.Data.Attrs.Name,
		Note: req.Data.Attrs.Note,
	})
	if err != nil {
		respErr := toRespErr(err)
		if errors.Is(err, ErrKindExists) {
			respErr.Source = &RespErrSource{Pointer: "/data/attributes/name"}
		}
		return nil, []RespErr{respErr}
	}

	out := fooToData(newFoo)
	return &out, nil
}

func (s *ServerV2) readFooV1(ctx context.Context, r *http.Request) (*Data[ResourceFooAttrs], []RespErr) {
	f, err := s.svc.ReadFoo(ctx, r.PathValue("id"))
	if err != nil {
		return nil, []RespErr{toRespErr(err)}
	}

	out := fooToData(f)
	return &out, nil
}

type (
	ReqUpdateFooV1 = ReqBody[FooUpdAttrs]

	FooUpdAttrs struct {
		Name *string `json:"name"`
		Note *string `json:"note"`
	}
)

func (s *ServerV2) updateFooV1(ctx context.Context, req ReqUpdateFooV1) (*Data[ResourceFooAttrs], []RespErr) {
	existing, err := s.svc.UpdateFoo(ctx, FooUpd{
		ID:   req.Data.ID,
		Name: req.Data.Attrs.Name,
		Note: req.Data.Attrs.Note,
	})
	if err != nil {
		respErr := toRespErr(err)
		if errors.Is(err, ErrKindExists) {
			respErr.Source = &RespErrSource{Pointer: "/data/attributes/name"}
		}
		return nil, []RespErr{respErr}
	}

	out := fooToData(existing)
	return &out, nil
}

func (s *ServerV2) delFooV1(ctx context.Context, r *http.Request) []RespErr {
	id := r.PathValue("id")
	if err := s.svc.DelFoo(ctx, id); err != nil {
		return []RespErr{toRespErr(err)}
	}
	return nil
}

func fooToData(f Foo) Data[ResourceFooAttrs] {
	return toFooData(f.ID, ResourceFooAttrs{
		Name:      f.Name,
		Note:      f.Note,
		CreatedAt: toTimestamp(f.CreatedAt),
		UpdatedAt: toTimestamp(f.UpdatedAt),
	})
}

func toFooData(id string, attrs ResourceFooAttrs) Data[ResourceFooAttrs] {
	return Data[ResourceFooAttrs]{
		Type:  resourceTypeFoo,
		ID:    id,
		Attrs: attrs,
	}
}

func toTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

func jsonIn[ReqAttr, RespAttr Attrs](resource string, successCode int, fn func(context.Context, ReqBody[ReqAttr]) (*Data[RespAttr], []RespErr)) http.Handler {
	return handler(successCode, func(ctx context.Context, r *http.Request) (*Data[RespAttr], []RespErr) {
		var reqBody ReqBody[ReqAttr]
		if respErr := decodeReq(r, &reqBody); respErr != nil {
			return nil, []RespErr{*respErr}
		}
		if reqBody.Data.Type != resource {
			return nil, []RespErr{{
				Status: http.StatusUnprocessableEntity,
				Code:   errCode(ErrKindInvalid),
				Msg:    "type must be " + resource,
				Source: &RespErrSource{
					Pointer: "/data/type",
				},
			}}
		}

		return fn(r.Context(), reqBody)
	})
}

func read[Attr any | []Attr](fn func(ctx context.Context, r *http.Request) (*Data[Attr], []RespErr)) http.Handler {
	return handler(http.StatusOK, fn)
}

func del(fn func(ctx context.Context, r *http.Request) []RespErr) http.Handler {
	return handler(http.StatusOK, func(ctx context.Context, r *http.Request) (*Data[any], []RespErr) {
		return nil, fn(ctx, r)
	})
}

func handler[Attr Attrs](successCode int, fn func(ctx context.Context, req *http.Request) (*Data[Attr], []RespErr)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, errs := fn(r.Context(), r)

		status := successCode
		for _, e := range errs {
			if e.Status > status {
				status = e.Status
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(RespBody[Attr]{
			Meta: getMeta(r.Context()),
			Errs: errs,
			Data: out,
		}) // 10.b)
	})
}

func decodeReq[Attr Attrs](r *http.Request, v *ReqBody[Attr]) *RespErr {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respErr := RespErr{
			Status: http.StatusBadRequest,
			Msg:    "failed to decode request body: " + err.Error(),
			Source: &RespErrSource{
				Pointer: "/data",
			},
			Code: errCode(ErrKindInvalid),
		}
		if unmarshErr := new(json.UnmarshalTypeError); errors.As(err, &unmarshErr) {
			respErr.Source.Pointer += "/data"
		}
		return &respErr
	}
	if r.Method == http.MethodPatch && r.PathValue("id") != v.Data.ID {
		return &RespErr{
			Status: http.StatusBadRequest,
			Msg:    "path id and data id must match",
			Source: &RespErrSource{
				Pointer: "/data/id",
			},
			Code: errCode(ErrKindInvalid),
		}
	}

	return nil
}

func toRespErr(err error) RespErr {
	return RespErr{
		Status: errStatus(err),
		Code:   errCode(err),
		Msg:    err.Error(),
	}
}

func errStatus(err error) int {
	switch {
	case errors.Is(err, ErrKindExists):
		return http.StatusConflict
	case errors.Is(err, ErrKindInvalid):
		return http.StatusBadRequest
	case errors.Is(err, ErrKindNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrKindUnAuthed):
		return http.StatusUnauthorized
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
					json.NewEncoder(w).Encode(RespBody[any]{
						Meta: getMeta(r.Context()),
						Errs: []RespErr{{
							Status: http.StatusUnauthorized,
							Code:   errCode(ErrKindUnAuthed),
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
			json.NewEncoder(w).Encode(RespBody[any]{
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
