package allsrv

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-metrics"
)

/*

	Concerns:
	✅1) the server depends on a hard type, coupling to the exact inmem db
		a) what happens if we want a different db?
	✅2) auth is copy-pasted in each handler
		a) what happens if we forget that copy pasta?
	✅3) auth is hardcoded to basic auth
		a) what happens if we want to adapt some other means of auth?
	✅4) router being used is the GLOBAL http.DefaultServeMux
		a) should avoid globals
		b) what happens if you have multiple servers in this go module who reference default serve mux?
	✅5) no tests
		a) how do we ensure things work?
		b) how do we know what is intended by the current implementation?
	6) http/db are coupled to the same type
		a) what happens when the concerns diverge? aka http wants a shape the db does not? (note: it happens A LOT)
	7) Server only works with HTTP
		a) what happens when we want to support grpc? thrift? other protocol?
		b) this setup often leads to copy pasta/weak abstractions that tend to leak
	✅8) Errors are opaque and limited
	9) API is very bare bones
		a) there is nothing actionable, so how does the consumer know to handle the error?
		b) if the APIs evolve, how does the consumer distinguish between old and new?
	10) Observability....
		✅a) metrics
		b) logging
		✅c) tracing
	✅11) hard coding UUID generation into db
	✅12) possible race conditions in inmem store
	✅13) there is a bug in the delete foo inmem db implementation

	Praises:
	1) minimal public API
	2) simple to read
	3) minimal indirection/obvious code
	4) is trivial in scope
*/

// Server dependencies
type (
	// DB represents the foo persistence layer.
	DB interface {
		CreateFoo(ctx context.Context, f Foo) error
		ReadFoo(ctx context.Context, id string) (Foo, error)
		UpdateFoo(ctx context.Context, f Foo) error
		DelFoo(ctx context.Context, id string) error
	}
)

type serverOpts struct {
	authFn func(http.Handler) http.Handler
	idFn   func() string
	nowFn  func() time.Time

	met *metrics.Metrics
	mux *http.ServeMux
}

// WithBasicAuth sets the authorization fn for the server to basic auth.
// 3)
func WithBasicAuth(user, pass string) func(*serverOpts) {
	return func(s *serverOpts) {
		s.authFn = basicAuth(user, pass)
	}
}

// WithIDFn sets the id generation fn for the server.
func WithIDFn(fn func() string) func(*serverOpts) {
	return func(s *serverOpts) {
		s.idFn = fn
	}
}

type Server struct {
	db  DB             // 1)
	mux *http.ServeMux // 4)

	authFn func(http.Handler) http.Handler // 3)
	idFn   func() string                   // 11)
}

func NewServer(db DB, opts ...func(*serverOpts)) *Server {
	opt := serverOpts{
		authFn: func(next http.Handler) http.Handler { // 3)
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// defaults to no auth
				next.ServeHTTP(w, r)
			})
		},
		idFn: func() string {
			// defaults to using a uuid
			return uuid.Must(uuid.NewV4()).String()
		},
		mux: http.NewServeMux(),
	}
	for _, o := range opts {
		o(&opt)
	}

	s := Server{
		db:     db,
		mux:    opt.mux, // 4)
		authFn: opt.authFn,
		idFn:   opt.idFn,
	}

	s.routes()
	return &s
}

func (s *Server) routes() {
	mw := applyMW(s.authFn, deprecationHeaders) // 2)

	// 4) 7) 9) 10)
	s.mux.Handle("POST /foo", mw(http.HandlerFunc(s.createFoo)))
	s.mux.Handle("GET /foo", mw(http.HandlerFunc(s.readFoo)))
	s.mux.Handle("PUT /foo", mw(http.HandlerFunc(s.updateFoo)))
	s.mux.Handle("DELETE /foo", mw(http.HandlerFunc(s.delFoo)))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 4)
	s.mux.ServeHTTP(w, r)
}

type Foo struct {
	// 6)
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (s *Server) createFoo(w http.ResponseWriter, r *http.Request) {
	var f Foo
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusForbidden) // 9)
		return
	}

	f.ID = s.idFn() // 11)

	if err := s.db.CreateFoo(r.Context(), f); err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 9)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Printf("unexpected error writing json value to response body: " + err.Error()) // 8) 10)
	}
}

func (s *Server) readFoo(w http.ResponseWriter, r *http.Request) {
	f, err := s.db.ReadFoo(r.Context(), r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound) // 9)
		return
	}

	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Printf("unexpected error writing json value to response body: " + err.Error()) // 8) 10)
	}
}

func (s *Server) updateFoo(w http.ResponseWriter, r *http.Request) {
	var f Foo
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusForbidden) // 9)
		return
	}

	if err := s.db.UpdateFoo(r.Context(), f); err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 9)
		return
	}
}

func (s *Server) delFoo(w http.ResponseWriter, r *http.Request) {
	if err := s.db.DelFoo(r.Context(), r.URL.Query().Get("id")); err != nil {
		w.WriteHeader(http.StatusNotFound) // 9)
		return
	}
}

// basicAuth provides a basic auth middleware to an http server.
// 2)
func basicAuth(expectedUser, expectedPass string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if user, pass, ok := r.BasicAuth(); !(ok && user == expectedUser && pass == expectedPass) {
				w.WriteHeader(http.StatusUnauthorized) // 9)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func deprecationHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", "Fri, 26 July 2024 23:59:59 GMT")
		next.ServeHTTP(w, r)
	})
}
