package allsrv

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	
	"github.com/gofrs/uuid"
)

/*

	Concerns:
	1) the server depends on a hard type, coupling to the exact inmem db
		a) what happens if we want a different db?
	2) auth is copy-pasted in each handler
		a) what happens if we forget that copy pasta?
	3) auth is hardcoded to basic auth
		a) what happens if we want to adapt some other means of auth?
	4) router being used is the GLOBAL http.DefaultServeMux
		a) should avoid globals
		b) what happens if you have multiple servers in this go module who reference default serve mux?
	5) no tests
		a) how do we ensure things work?
		b) how do we know what is intended by the current implementation?
	6) http/db are coupled to the same type
		a) what happens when the concerns diverge? aka http wants a shape the db does not? (note: it happens A LOT)
	7) Server only works with HTTP
		a) what happens when we want to support grpc? thrift? other protocol?
		b) this setup often leads to copy pasta/weak abstractions that tend to leak
	8) Errors are opaque and limited
	9) API is very bare bones
		a) there is nothing actionable, so how does the consumer know to handle the error?
		b) if the APIs evolve, how does the consumer distinguish between old and new?
	10) Observability....
	11) hard coding UUID generation into db
	12) possible race conditions in inmem store

	Praises:
	1) minimal public API
	2) simple to read
	3) minimal indirection/obvious code
	4) is trivial in scope
*/

type Server struct {
	db *InmemDB // 1)
	
	user, pass string // 3)
}

func NewServer(db *InmemDB, user, pass string) *Server {
	s := Server{
		db:   db,
		user: user,
		pass: pass,
	}
	s.routes()
	return &s
}

func (s *Server) routes() {
	// 4) 7) 9) 10)
	http.Handle("POST /foo", http.HandlerFunc(s.createFoo))
	http.Handle("GET /foo", http.HandlerFunc(s.readFoo))
	http.Handle("PUT /foo", http.HandlerFunc(s.updateFoo))
	http.Handle("DELETE /foo", http.HandlerFunc(s.delFoo))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 4)
	http.DefaultServeMux.ServeHTTP(w, r)
}

type Foo struct {
	// 6)
	ID   string `json:"id" gorm:"id"`
	Name string `json:"name" gorm:"name"`
	Note string `json:"note" gorm:"note"`
}

func (s *Server) createFoo(w http.ResponseWriter, r *http.Request) {
	// 2)
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized) // 9)
		return
	}
	
	var f Foo
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusForbidden) // 9)
		return
	}
	
	newFooID, err := s.db.createFoo(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 9)
		return
	}
	
	f, err = s.db.readFoo(newFooID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) // 9)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Printf("unexpected error writing json value to response body: " + err.Error()) // 8) 10)
	}
}

func (s *Server) readFoo(w http.ResponseWriter, r *http.Request) {
	// 2)
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized) // 9)
		return
	}
	
	f, err := s.db.readFoo(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound) // 9)
		return
	}
	
	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Printf("unexpected error writing json value to response body: " + err.Error()) // 8) 10)
	}
}

func (s *Server) updateFoo(w http.ResponseWriter, r *http.Request) {
	// 2)
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized) // 9)
		return
	}
	
	var f Foo
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusForbidden) // 9)
		return
	}
	
	if err := s.db.updateFoo(f); err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 9)
		return
	}
}

func (s *Server) delFoo(w http.ResponseWriter, r *http.Request) {
	// 2)
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized) // 9)
		return
	}
	
	if err := s.db.delFoo(r.URL.Query().Get("id")); err != nil {
		w.WriteHeader(http.StatusNotFound) // 9)
		return
	}
}

// InmemDB is an in-memory store.
type InmemDB struct {
	m []Foo // 12)
}

func (db *InmemDB) createFoo(f Foo) (string, error) {
	f.ID = uuid.Must(uuid.NewV4()).String() // 11)
	
	for _, existing := range db.m {
		if f.Name == existing.Name {
			return "", errors.New("foo " + f.Name + " exists") // 8)
		}
	}
	
	db.m = append(db.m, f)
	
	return f.ID, nil
}

func (db *InmemDB) readFoo(id string) (Foo, error) {
	for _, f := range db.m {
		if id == f.ID {
			return f, nil
		}
	}
	return Foo{}, errors.New("foo not found for id: " + id) // 8)
}

func (db *InmemDB) updateFoo(f Foo) error {
	for i, existing := range db.m {
		if f.ID == existing.ID {
			db.m[i] = f
			return nil
		}
	}
	return errors.New("foo not found for id: " + f.ID) // 8)
}

func (db *InmemDB) delFoo(id string) error {
	for i, f := range db.m {
		if id == f.ID {
			db.m = append(db.m[:i], db.m[i+1:]...)
		}
	}
	return errors.New("foo not found for id: " + id) // 8)
}
