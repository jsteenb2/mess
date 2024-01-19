package allsrv

import (
	"context"
	"sync"
)

// InmemDB is an in-memory store.
type InmemDB struct {
	mu sync.Mutex
	m  []Foo // 12)
}

func (db *InmemDB) CreateFoo(_ context.Context, f Foo) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, existing := range db.m {
		if f.Name == existing.Name {
			return ExistsErr("foo "+f.Name+" exists", "name", f.Name, "existing_foo_id", existing.ID) // 8)
		}
	}

	db.m = append(db.m, f)

	return nil
}

func (db *InmemDB) ReadFoo(_ context.Context, id string) (Foo, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, f := range db.m {
		if id == f.ID {
			return f, nil
		}
	}
	return Foo{}, NotFoundErr("foo not found for id: "+id, "id", id) // 8)
}

func (db *InmemDB) UpdateFoo(_ context.Context, f Foo) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, existing := range db.m {
		if existing.Name == f.Name && existing.ID != f.ID {
			return ExistsErr("foo "+f.Name+" exists", "name", f.Name, "existing_foo_id", existing.ID) // 8)
		}
	}

	for i, existing := range db.m {
		if f.ID == existing.ID {
			db.m[i] = f
			return nil
		}
	}
	return NotFoundErr("foo not found for id: "+f.ID, "id", f.ID) // 8)
}

func (db *InmemDB) DelFoo(_ context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i, f := range db.m {
		if id == f.ID {
			db.m = append(db.m[:i], db.m[i+1:]...)
			return nil // 13)
		}
	}
	return NotFoundErr("foo not found for id: "+id, "id", id) // 8)
}
