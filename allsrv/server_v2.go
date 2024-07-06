package allsrv

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-metrics"
	"github.com/jsteenb2/errors"
	
	"github.com/jsteenb2/allsrvc"
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
	mw = append(mw, withOriginUserAgent, withTraceID, withStartTime)
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

const (
	resourceTypeFoo = "foo"
)

func (s *ServerV2) createFooV1(ctx context.Context, req allsrvc.ReqBody[allsrvc.FooCreateAttrs]) (*allsrvc.Data[allsrvc.ResourceFooAttrs], []allsrvc.RespErr) {
	newFoo, err := s.svc.CreateFoo(ctx, Foo{
		Name: req.Data.Attrs.Name,
		Note: req.Data.Attrs.Note,
	})
	if err != nil {
		respErr := toRespErr(err)
		if errors.Is(err, ErrKindExists) {
			respErr.Source = &allsrvc.RespErrSource{Pointer: "/data/attributes/name"}
		}
		return nil, []allsrvc.RespErr{respErr}
	}
	
	out := FooToData(newFoo)
	return &out, nil
}

func (s *ServerV2) readFooV1(ctx context.Context, r *http.Request) (*allsrvc.Data[allsrvc.ResourceFooAttrs], []allsrvc.RespErr) {
	f, err := s.svc.ReadFoo(ctx, r.PathValue("id"))
	if err != nil {
		return nil, []allsrvc.RespErr{toRespErr(err)}
	}
	
	out := FooToData(f)
	return &out, nil
}

func (s *ServerV2) updateFooV1(ctx context.Context, req allsrvc.ReqBody[allsrvc.FooUpdAttrs]) (*allsrvc.Data[allsrvc.ResourceFooAttrs], []allsrvc.RespErr) {
	existing, err := s.svc.UpdateFoo(ctx, FooUpd{
		ID:   req.Data.ID,
		Name: req.Data.Attrs.Name,
		Note: req.Data.Attrs.Note,
	})
	if err != nil {
		respErr := toRespErr(err)
		if errors.Is(err, ErrKindExists) {
			respErr.Source = &allsrvc.RespErrSource{Pointer: "/data/attributes/name"}
		}
		return nil, []allsrvc.RespErr{respErr}
	}
	
	out := FooToData(existing)
	return &out, nil
}

func (s *ServerV2) delFooV1(ctx context.Context, r *http.Request) []allsrvc.RespErr {
	id := r.PathValue("id")
	if err := s.svc.DelFoo(ctx, id); err != nil {
		return []allsrvc.RespErr{toRespErr(err)}
	}
	return nil
}

func FooToData(f Foo) allsrvc.Data[allsrvc.ResourceFooAttrs] {
	return allsrvc.Data[allsrvc.ResourceFooAttrs]{
		Type: resourceTypeFoo,
		ID:   f.ID,
		Attrs: allsrvc.ResourceFooAttrs{
			Name:      f.Name,
			Note:      f.Note,
			CreatedAt: toTimestamp(f.CreatedAt),
			UpdatedAt: toTimestamp(f.UpdatedAt),
		},
	}
}

func toTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

func jsonIn[ReqAttr, RespAttr allsrvc.Attrs](
	resource string,
	successCode int,
	fn func(context.Context, allsrvc.ReqBody[ReqAttr]) (*allsrvc.Data[RespAttr], []allsrvc.RespErr),
) http.Handler {
	return handler(successCode, func(ctx context.Context, r *http.Request) (*allsrvc.Data[RespAttr], []allsrvc.RespErr) {
		var reqBody allsrvc.ReqBody[ReqAttr]
		if respErr := decodeReq(r, &reqBody); respErr != nil {
			return nil, []allsrvc.RespErr{*respErr}
		}
		if reqBody.Data.Type != resource {
			return nil, []allsrvc.RespErr{{
				Status: http.StatusUnprocessableEntity,
				Code:   errCode(ErrKindInvalid),
				Msg:    "type must be " + resource,
				Source: &allsrvc.RespErrSource{
					Pointer: "/data/type",
				},
			}}
		}
		
		return fn(r.Context(), reqBody)
	})
}

func read[Attr any | []Attr](fn func(ctx context.Context, r *http.Request) (*allsrvc.Data[Attr], []allsrvc.RespErr)) http.Handler {
	return handler(http.StatusOK, fn)
}

func del(fn func(ctx context.Context, r *http.Request) []allsrvc.RespErr) http.Handler {
	return handler(http.StatusOK, func(ctx context.Context, r *http.Request) (*allsrvc.Data[any], []allsrvc.RespErr) {
		return nil, fn(ctx, r)
	})
}

func handler[Attr allsrvc.Attrs](successCode int, fn func(ctx context.Context, req *http.Request) (*allsrvc.Data[Attr], []allsrvc.RespErr)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, errs := fn(r.Context(), r)
		
		status := successCode
		for _, e := range errs {
			if e.Status > status {
				status = e.Status
			}
		}
		writeResp(w, status, allsrvc.RespBody[Attr]{
			Meta: getMeta(r.Context()),
			Errs: errs,
			Data: out,
		}) // 10.b)
	})
}

func writeResp(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) // 10.b)
}

func decodeReq[Attr allsrvc.Attrs](r *http.Request, v *allsrvc.ReqBody[Attr]) *allsrvc.RespErr {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respErr := allsrvc.RespErr{
			Status: http.StatusBadRequest,
			Msg:    "failed to decode request body: " + err.Error(),
			Source: &allsrvc.RespErrSource{
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
		return &allsrvc.RespErr{
			Status: http.StatusBadRequest,
			Msg:    "path id and data id must match",
			Source: &allsrvc.RespErrSource{
				Pointer: "/data/id",
			},
			Code: errCode(ErrKindInvalid),
		}
	}
	
	return nil
}

func toRespErr(err error) allsrvc.RespErr {
	return allsrvc.RespErr{
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
					writeResp(w, http.StatusUnauthorized, allsrvc.RespBody[any]{
						Meta: getMeta(r.Context()),
						Errs: []allsrvc.RespErr{{
							Status: http.StatusUnauthorized,
							Code:   errCode(ErrKindUnAuthed),
							Msg:    "unauthorized access",
							Source: &allsrvc.RespErrSource{
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
			writeResp(w, http.StatusUnsupportedMediaType, allsrvc.RespBody[any]{
				Meta: getMeta(r.Context()),
				Errs: []allsrvc.RespErr{{
					Code: http.StatusUnsupportedMediaType,
					Msg:  "received invalid media type",
				}},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getMeta(ctx context.Context) allsrvc.RespMeta {
	return allsrvc.RespMeta{
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

type ctxKey string

const (
	ctxKeyOrigin    ctxKey = "origin"
	ctxStartTime    ctxKey = "start"
	ctxTraceID      ctxKey = "trace-id"
	ctxKeyUserAgent ctxKey = "user_agent"
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

func withStartTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxStartTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withOriginUserAgent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKeyOrigin, r.Header.Get("Origin"))
		ctx = context.WithValue(ctx, ctxKeyUserAgent, r.Header.Get("User-Agent"))
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value(ctxTraceID).(string)
	return traceID
}

func getOrigin(ctx context.Context) string {
	origin, _ := ctx.Value(ctxKeyOrigin).(string)
	return origin
}

func getUserAgent(ctx context.Context) string {
	userAgent, _ := ctx.Value(ctxKeyUserAgent).(string)
	return userAgent
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
