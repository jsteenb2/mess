package allsrv

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	
	"github.com/gofrs/uuid"
)

type Server struct {
	db *inmemDB
	
	user, pass string
}

func NewServer(db *inmemDB, user, pass string) *Server {
	s := Server{
		db:   db,
		user: user,
		pass: pass,
	}
	s.routes()
	return &s
}

func (s *Server) routes() {
	http.Handle("POST /foo", http.HandlerFunc(s.createFoo))
	http.Handle("GET /foo", http.HandlerFunc(s.readFoo))
	http.Handle("PUT /foo", http.HandlerFunc(s.updateFoo))
	http.Handle("DELETE /foo", http.HandlerFunc(s.delFoo))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, r)
}

type Foo struct {
	ID   string `json:"id" gorm:"id"`
	Name string `json:"name" gorm:"name"`
	Note string `json:"note" gorm:"note"`
}

func (s *Server) createFoo(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	
	var f Foo
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	
	newFooID, err := s.db.createFoo(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	f, err = s.db.readFoo(newFooID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Printf("unexpected error writing json value to response body: " + err.Error())
	}
}

func (s *Server) readFoo(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	
	f, err := s.db.readFoo(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Printf("unexpected error writing json value to response body: " + err.Error())
	}
}

func (s *Server) updateFoo(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	
	var f Foo
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	
	if err := s.db.updateFoo(f); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) delFoo(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !(ok && user == s.user && pass == s.pass) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	
	if err := s.db.delFoo(r.URL.Query().Get("id")); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

type inmemDB struct {
	m []Foo
}

func (db *inmemDB) createFoo(f Foo) (string, error) {
	f.ID = uuid.Must(uuid.NewV4()).String()
	
	for _, existing := range db.m {
		if f.Name == existing.Name {
			return "", errors.New("foo " + f.Name + " exists")
		}
	}
	
	db.m = append(db.m, f)
	
	return f.ID, nil
}

func (db *inmemDB) readFoo(id string) (Foo, error) {
	for _, f := range db.m {
		if id == f.ID {
			return f, nil
		}
	}
	return Foo{}, errors.New("foo not found for id: " + id)
}

func (db *inmemDB) updateFoo(f Foo) error {
	for i, existing := range db.m {
		if f.ID == existing.ID {
			db.m[i] = f
			return nil
		}
	}
	return errors.New("foo not found for id: " + f.ID)
}

func (db *inmemDB) delFoo(id string) error {
	for i, f := range db.m {
		if id == f.ID {
			db.m = append(db.m[:i], db.m[i+1:]...)
		}
	}
	return errors.New("foo not found for id: " + id)
}
